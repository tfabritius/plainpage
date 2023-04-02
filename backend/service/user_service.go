package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tfabritius/plainpage/libs/argon2"
	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/storage"
)

func NewUserService(s storage.Storage) UserService {
	return UserService{
		storage: s,
	}
}

type UserService struct {
	storage storage.Storage
}

func (s *UserService) SetUsername(user *storage.User, username string) error {
	if !s.isValidUsername(username) {
		return storage.ErrInvalidUsername
	}
	user.Username = username
	return nil
}

func (*UserService) SetPasswordHash(user *storage.User, password string) error {
	hash, err := argon2.HashPasswordDefault(password)
	if err != nil {
		return err
	}
	user.PasswordHash = "argon2:" + hash
	return nil
}

func (s *UserService) Create(username, password, realName string) (storage.User, error) {

	users, err := s.storage.GetAllUsers()
	if err != nil {
		return storage.User{}, err
	}

	if !s.isUsernameUnique(users, username) {
		return storage.User{}, storage.ErrUserExistsAlready
	}

	id, err := utils.GenerateRandomString(6)
	if err != nil {
		return storage.User{}, err
	}

	user := storage.User{
		ID:       id,
		RealName: realName,
	}

	if err := s.SetUsername(&user, username); err != nil {
		return storage.User{}, err
	}

	if err := s.SetPasswordHash(&user, password); err != nil {
		return storage.User{}, err
	}

	users = append(users, user)

	err = s.storage.SaveAllUsers(users)
	if err != nil {
		return storage.User{}, err
	}

	return user, nil
}

func (s *UserService) GetByUsername(username string) (storage.User, error) {
	users, err := s.storage.GetAllUsers()
	if err != nil {
		return storage.User{}, fmt.Errorf("could not read users: %w", err)
	}

	for _, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return user, nil
		}
	}

	return storage.User{}, storage.ErrNotFound
}

func (*UserService) getById(users []storage.User, id string) *storage.User {
	for i := range users {
		if users[i].ID == id {
			return &users[i]
		}
	}

	return nil
}

func (s *UserService) Save(user storage.User) error {
	users, err := s.storage.GetAllUsers()
	if err != nil {
		return fmt.Errorf("could not read users: %w", err)
	}

	existingUser := s.getById(users, user.ID)
	if existingUser == nil {
		return storage.ErrNotFound
	}

	if user.Username != existingUser.Username {
		if !s.isUsernameUnique(users, user.Username) {
			return storage.ErrUserExistsAlready
		}
	}

	existingUser.Username = user.Username
	existingUser.RealName = user.RealName
	existingUser.PasswordHash = user.PasswordHash

	if err := s.storage.SaveAllUsers(users); err != nil {
		return fmt.Errorf("could not save users: %w", err)
	}

	return nil
}

func (s *UserService) DeleteByUsername(username string) error {
	users, err := s.storage.GetAllUsers()
	if err != nil {
		return fmt.Errorf("could not read users: %w", err)
	}

	found := false

	for i := 0; i < len(users); {
		if strings.ToLower(users[i].Username) == strings.ToLower(username) {
			found = true
			users = append(users[:i], users[i+1:]...)
		} else {
			i++
		}
	}
	if !found {
		return storage.ErrNotFound
	}

	if err := s.storage.SaveAllUsers(users); err != nil {
		return fmt.Errorf("could not save users: %w", err)
	}

	return nil
}

func (s *UserService) EnhanceACLsWithUserInfo(meta *storage.PageMeta) error {
	if meta.ACLs != nil {
		users, err := s.storage.GetAllUsers()
		if err != nil {
			return fmt.Errorf("could not read users: %w", err)
		}

		for i, acl := range *meta.ACLs {
			if userId, found := strings.CutPrefix(acl.Subject, "user:"); found {
				user := s.getById(users, userId)
				(*meta.ACLs)[i].User = user
			}
		}
	}

	return nil
}

func (*UserService) isUsernameUnique(users []storage.User, username string) bool {
	for _, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return false
		}
	}
	return true
}

func (*UserService) isValidUsername(username string) bool {
	regex := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_\\.-]{3,20}$")
	return regex.MatchString(username)
}
