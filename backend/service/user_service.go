package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/tfabritius/plainpage/libs/argon2"
	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/model"
	"gopkg.in/yaml.v3"
)

func NewUserService(store model.Storage) UserService {
	s := UserService{
		storage: store,
	}

	// Initialize users.yml
	if !s.storage.Exists("users.yml") {
		err := s.saveAll([]model.User{})
		if err != nil {
			log.Fatalln("Could not create users.yml:", err)
		}
	}

	return s
}

type UserService struct {
	storage model.Storage
}

func (s *UserService) ReadAll() ([]model.User, error) {
	bytes, err := s.storage.ReadFile("users.yml")
	if err != nil {
		return nil, fmt.Errorf("could not read users.yml: %w", err)
	}

	// parse YAML
	users := []model.User{}
	if err := yaml.Unmarshal(bytes, &users); err != nil {
		return nil, fmt.Errorf("could not parse YAML: %w", err)
	}

	return users, nil
}

func (s *UserService) saveAll(users []model.User) error {
	bytes, err := yaml.Marshal(&users)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := s.storage.WriteFile("users.yml", bytes); err != nil {
		return fmt.Errorf("could not write users.yml: %w", err)
	}

	return nil
}

func (s *UserService) SetUsername(user *model.User, username string) error {
	if !s.isValidUsername(username) {
		return model.ErrInvalidUsername
	}
	user.Username = username
	return nil
}

func (*UserService) SetPasswordHash(user *model.User, password string) error {
	hash, err := argon2.HashPasswordDefault(password)
	if err != nil {
		return err
	}
	user.PasswordHash = "argon2:" + hash
	return nil
}

func (*UserService) verifyPassword(user model.User, password string) bool {
	if plain, found := strings.CutPrefix(user.PasswordHash, "plain:"); found {
		return password == plain
	}

	if argon2Hash, found := strings.CutPrefix(user.PasswordHash, "argon2:"); found {
		match, err := argon2.VerifyPassword(password, argon2Hash)
		if err != nil {
			return false
		}
		return match
	}

	return false
}

func (s *UserService) Create(username, password, displayName string) (model.User, error) {

	users, err := s.ReadAll()
	if err != nil {
		return model.User{}, err
	}

	if !s.isUsernameUnique(users, username) {
		return model.User{}, model.ErrUserExistsAlready
	}

	id, err := utils.GenerateRandomString(6)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		ID:          id,
		DisplayName: displayName,
	}

	if err := s.SetUsername(&user, username); err != nil {
		return model.User{}, err
	}

	if err := s.SetPasswordHash(&user, password); err != nil {
		return model.User{}, err
	}

	users = append(users, user)

	err = s.saveAll(users)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *UserService) GetByUsername(username string) (model.User, error) {
	users, err := s.ReadAll()
	if err != nil {
		return model.User{}, fmt.Errorf("could not read users: %w", err)
	}

	for _, user := range users {
		if strings.EqualFold(user.Username, username) {
			return user, nil
		}
	}

	return model.User{}, model.ErrNotFound
}

func (s *UserService) GetById(id string) (model.User, error) {
	users, err := s.ReadAll()
	if err != nil {
		return model.User{}, fmt.Errorf("could not read users: %w", err)
	}

	user := s.filterById(users, id)
	if user != nil {
		return *user, nil
	}

	return model.User{}, model.ErrNotFound
}

func (s *UserService) VerifyCredentials(username, password string) (*model.User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Verify the provided password against the stored hash
	if !s.verifyPassword(user, password) {
		return nil, nil
	}

	// Hash plain passwords
	if strings.HasPrefix(user.PasswordHash, "plain:") {
		// Re-hash with argon2 and persist
		if err := s.SetPasswordHash(&user, password); err != nil {
			return nil, err
		}
		if err := s.Save(user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (*UserService) filterById(users []model.User, id string) *model.User {
	for i := range users {
		if users[i].ID == id {
			return &users[i]
		}
	}

	return nil
}

func (s *UserService) Save(user model.User) error {
	users, err := s.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read users: %w", err)
	}

	existingUser := s.filterById(users, user.ID)
	if existingUser == nil {
		return model.ErrNotFound
	}

	if user.Username != existingUser.Username {
		if !s.isUsernameUnique(users, user.Username) {
			return model.ErrUserExistsAlready
		}
	}

	existingUser.Username = user.Username
	existingUser.DisplayName = user.DisplayName
	existingUser.PasswordHash = user.PasswordHash

	if err := s.saveAll(users); err != nil {
		return fmt.Errorf("could not save users: %w", err)
	}

	return nil
}

func (s *UserService) DeleteByUsername(username string) error {
	users, err := s.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read users: %w", err)
	}

	found := false

	for i := 0; i < len(users); {
		if strings.EqualFold(users[i].Username, username) {
			found = true
			users = append(users[:i], users[i+1:]...)
		} else {
			i++
		}
	}
	if !found {
		return model.ErrNotFound
	}

	if err := s.saveAll(users); err != nil {
		return fmt.Errorf("could not save users: %w", err)
	}

	return nil
}

func (s *UserService) EnhanceACLWithUserInfo(acl *[]model.AccessRule) error {
	if acl != nil {
		users, err := s.ReadAll()
		if err != nil {
			return fmt.Errorf("could not read users: %w", err)
		}

		for i, rule := range *acl {
			if userId, found := strings.CutPrefix(rule.Subject, "user:"); found {
				user := s.filterById(users, userId)
				(*acl)[i].User = user
			}
		}
	}

	return nil
}

type AccessDeniedError struct {
	StatusCode int
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied (%v)", e.StatusCode)
}

func (s *UserService) CheckAppPermissions(
	userID string,
	op model.AccessOp,
) error {
	cfg, err := s.storage.ReadConfig()
	if err != nil {
		return err
	}

	return s.checkPermissions(cfg.ACL, userID, op, true)
}

func (s *UserService) CheckContentPermissions(
	acl *[]model.AccessRule,
	userID string,
	op model.AccessOp,
) error {
	return s.checkPermissions(*acl, userID, op, false)
}

func (s *UserService) checkPermissions(acl []model.AccessRule, userID string, op model.AccessOp, aclIsApp bool) error {

	// Allow access if anonymous is allowed
	if s.compareACL(acl, "anonymous", op) {
		return nil
	}

	// Deny anonymous access
	if userID == "" {
		return &AccessDeniedError{
			StatusCode: http.StatusUnauthorized,
		}
	}

	// Allow if all users are allowed
	if s.compareACL(acl, "all", op) {
		return nil
	}

	// Allow if user is allowed
	if s.compareACL(acl, "user:"+userID, op) {
		return nil
	}

	// Read global ACL
	if !aclIsApp {
		cfg, err := s.storage.ReadConfig()
		if err != nil {
			return err
		}
		acl = cfg.ACL
	}

	// Allow if user has admin privileges
	if s.compareACL(acl, "user:"+userID, model.AccessOpAdmin) {
		return nil
	}

	// Deny access
	return &AccessDeniedError{
		StatusCode: http.StatusForbidden,
	}
}

func (*UserService) compareACL(acl []model.AccessRule, subject string, op model.AccessOp) bool {
	for _, rule := range acl {
		if rule.Subject == subject {
			for _, o := range rule.Operations {
				if o == op {
					return true
				}
			}
			return false
		}
	}
	return false
}

func (*UserService) isUsernameUnique(users []model.User, username string) bool {
	for _, user := range users {
		if strings.EqualFold(user.Username, username) {
			return false
		}
	}
	return true
}

func (*UserService) isValidUsername(username string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\.-]{3,20}$`)
	return regex.MatchString(username)
}
