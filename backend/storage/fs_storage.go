package storage

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tfabritius/plainpage/libs/utils"
	"gopkg.in/yaml.v3"
)

type fsStorage struct {
	DataDir string
}

func touch(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	return err
}

func NewFsStorage(dataDir string) Storage {
	log.Println("Data directory:", dataDir)

	fi, err := os.Stat(dataDir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			log.Fatalln("Could not create data directory:", err)
		}
		log.Println("Data directory created")
	} else if err != nil {
		log.Fatalln("Cannot access data directory:", err)
	} else if !fi.IsDir() {
		log.Fatalln("Data directory is not a directory")
	}

	storage := fsStorage{DataDir: dataDir}

	// Create pages and attic directories
	for _, folder := range []string{"pages", "attic"} {
		// Create directory, continue if it exists already
		err := os.MkdirAll(filepath.Join(storage.DataDir, folder), 0755)
		if err != nil {
			log.Fatalln("Could not create "+folder+" folder:", err)
		}
	}

	// Create configuration file
	touch(filepath.Join(storage.DataDir, "users.yml"))

	return &storage
}

func (fss *fsStorage) getFsPathOfPage(urlPath string) string {
	return filepath.Join(fss.DataDir, "pages", urlPath+".md")
}

func (fss *fsStorage) getFsPathOfAtticPage(urlPath string, revision int64) string {
	revStr := strconv.FormatInt(revision, 10)
	return filepath.Join(fss.DataDir, "attic", urlPath+"."+revStr+".md")
}

func (fss *fsStorage) getFsPathOfFolder(urlPath string) string {
	return filepath.Join(fss.DataDir, "pages", urlPath)
}

func (fss *fsStorage) IsPage(urlPath string) bool {
	fsPath := fss.getFsPathOfPage(urlPath)
	_, err := os.Stat(fsPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (fss *fsStorage) IsAtticPage(urlPath string, revision int64) bool {
	fsPath := fss.getFsPathOfAtticPage(urlPath, revision)
	_, err := os.Stat(fsPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (fss *fsStorage) IsFolder(urlPath string) bool {
	fsPath := fss.getFsPathOfFolder(urlPath)
	_, err := os.Stat(fsPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (fss *fsStorage) CreateFolder(urlPath string) error {
	if !fss.IsFolder(path.Dir(urlPath)) {
		return ErrParentFolderNotFound
	}
	if fss.IsPage(urlPath) || fss.IsFolder(urlPath) {
		return ErrPageOrFolderExistsAlready
	}

	fsPath := fss.getFsPathOfFolder(urlPath)
	return os.Mkdir(fsPath, 0755)

}

func (fss *fsStorage) SavePage(urlPath, content string, meta PageMeta) error {
	if !fss.IsFolder(path.Dir(urlPath)) {
		return ErrParentFolderNotFound
	}
	if fss.IsFolder(urlPath) {
		return ErrPageOrFolderExistsAlready
	}

	fsPath := fss.getFsPathOfPage(urlPath)

	serializedPage, err := serializeFrontMatter(meta, content)
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	if err := os.WriteFile(fsPath, []byte(serializedPage), 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	err = fss.savePageToAttic(urlPath, serializedPage)
	if err != nil {
		return fmt.Errorf("could not save page to attic: %w", err)
	}

	return nil
}

func (fss *fsStorage) DeletePage(urlPath string) error {
	fsPath := fss.getFsPathOfPage(urlPath)

	err := os.Remove(fsPath)
	if err != nil {
		return fmt.Errorf("could not remove file: %w", err)
	}
	return nil
}

func (fss *fsStorage) DeleteEmptyFolder(urlPath string) error {
	fsPath := fss.getFsPathOfFolder(urlPath)

	err := os.Remove(fsPath)
	if err != nil && strings.HasSuffix(err.Error(), "The directory is not empty.") {
		return ErrFolderNotEmpty
	}
	return err
}

func (fss *fsStorage) ReadPage(urlPath string, revision *int64) (Page, error) {
	var fsPath string
	if revision == nil {
		fsPath = fss.getFsPathOfPage(urlPath)
	} else {
		fsPath = fss.getFsPathOfAtticPage(urlPath, *revision)
	}

	// read the file's content
	bytes, err := os.ReadFile(fsPath)
	if err != nil {
		return Page{}, fmt.Errorf("could not read file: %w", err)
	}

	fm, content, err := parseFrontMatter(string(bytes))
	if err != nil {
		return Page{}, fmt.Errorf("could not parse frontmatter: %w", err)
	}

	// enhance ACLs with additional user information
	if fm.ACLs != nil {
		for i, acl := range *fm.ACLs {
			if userId, found := strings.CutPrefix(acl.Subject, "user:"); found {
				user, err := fss.getUserById(userId)
				if errors.Is(err, ErrNotFound) {
					continue
				}
				if err != nil {
					return Page{}, fmt.Errorf("failed to find user: %w", err)
				}

				(*fm.ACLs)[i].User = &user
			}
		}
	}

	u, err := url.JoinPath("/", urlPath)
	if err != nil {
		return Page{}, fmt.Errorf("could not join url: %w", err)
	}

	// create the response
	page := Page{
		Url:     u,
		Content: content,
		Meta:    fm,
	}
	return page, nil
}

func (fss *fsStorage) ReadFolder(urlPath string) ([]FolderEntry, error) {
	dirPath := fss.getFsPathOfFolder(urlPath)

	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("could not open directory: %w", err)
	}
	defer dir.Close()

	// Get a list of all files in the directory
	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	folderEntries := make([]FolderEntry, 0, len(fileInfos))
	for _, fi := range fileInfos {

		u, err := url.JoinPath("/", urlPath, fi.Name())
		if err != nil {
			return nil, fmt.Errorf("could not join url: %w", err)
		}

		e := FolderEntry{
			Url:      u,
			Name:     fi.Name(),
			IsFolder: fi.IsDir(),
		}
		if !e.IsFolder {
			if strings.HasSuffix(e.Name, ".md") {
				e.Name = strings.TrimSuffix(e.Name, ".md")
				e.Url = strings.TrimSuffix(e.Url, ".md")
			} else {
				continue
			}
		}

		folderEntries = append(folderEntries, e)
	}

	return folderEntries, nil
}

func (fss *fsStorage) ListAttic(urlPath string) ([]AtticEntry, error) {
	pageName := path.Base(urlPath)
	parentDir := filepath.Dir(fss.getFsPathOfAtticPage(urlPath, 0))

	// Open the directory
	dir, err := os.Open(parentDir)
	if err != nil {
		return nil, fmt.Errorf("could not open directory: %w", err)
	}
	defer dir.Close()

	// Get a list of all files in the directory
	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	atticEntries := []AtticEntry{}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}

		// Check if name start with page name
		name, found := strings.CutPrefix(fi.Name(), pageName+".")
		if !found {
			continue
		}

		// Check if name end with file extension
		name, found = strings.CutSuffix(name, ".md")
		if !found {
			continue
		}

		rev, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			continue
		}

		atticEntries = append(atticEntries, AtticEntry{Revision: rev})
	}

	return atticEntries, nil
}

func (fss *fsStorage) createDir(file string) error {
	dir := filepath.Dir(file)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("could not create directories: %w", err)
	}

	return nil
}

// savePageToAttic saves serialized page to attic directory
func (fss *fsStorage) savePageToAttic(urlPath string, serializedPage string) error {
	timestampInt := time.Now().Unix()
	atticFile := fss.getFsPathOfAtticPage(urlPath, timestampInt)

	// creates folders in atticPath
	if err := fss.createDir(atticFile); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	// write the file's content
	if err := os.WriteFile(atticFile, []byte(serializedPage), 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

func (fss *fsStorage) GetAllUsers() ([]User, error) {
	fsPath := filepath.Join(fss.DataDir, "users.yml")

	// read the file
	bytes, err := os.ReadFile(fsPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	// parse YAML
	users := []User{}
	if err := yaml.Unmarshal(bytes, &users); err != nil {
		return nil, fmt.Errorf("could not parse YAML: %w", err)
	}

	return users, nil
}

func (fss *fsStorage) getUserById(id string) (User, error) {
	users, err := fss.GetAllUsers()
	if err != nil {
		return User{}, fmt.Errorf("could not read users: %w", err)
	}

	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (fss *fsStorage) GetUserByUsername(username string) (User, error) {
	users, err := fss.GetAllUsers()
	if err != nil {
		return User{}, fmt.Errorf("could not read users: %w", err)
	}

	for _, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (fss *fsStorage) SaveAllUsers(users []User) error {
	fsPath := filepath.Join(fss.DataDir, "users.yml")

	bytes, err := yaml.Marshal(&users)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.WriteFile(fsPath, bytes, 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

func isValidUsername(username string) bool {
	regex := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_\\.-]{3,20}$")
	return regex.MatchString(username)
}

func (fss *fsStorage) AddUser(username, password, realName string) (User, error) {
	users, err := fss.GetAllUsers()
	if err != nil {
		return User{}, err
	}

	// make sure username only contains allowed characters
	if !isValidUsername(username) {
		return User{}, ErrInvalidUsername
	}

	// make sure (lowercase) username is unique
	for _, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return User{}, ErrUserExistsAlready
		}
	}

	id, err := utils.GenerateRandomString(6)
	if err != nil {
		return User{}, err
	}

	passwordHash := "plain:" + password

	user := User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		RealName:     realName,
	}

	users = append(users, user)

	err = fss.SaveAllUsers(users)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
