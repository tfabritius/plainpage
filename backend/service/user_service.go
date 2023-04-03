package service

import (
	"errors"
	"fmt"
	"net/http"
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

func (*UserService) verifyPassword(user storage.User, password string) bool {
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

func (s *UserService) GetById(id string) (storage.User, error) {
	users, err := s.storage.GetAllUsers()
	if err != nil {
		return storage.User{}, fmt.Errorf("could not read users: %w", err)
	}

	user := s.filterById(users, id)
	if user != nil {
		return *user, nil
	}

	return storage.User{}, storage.ErrNotFound
}

func (s *UserService) VerifyCredentials(username, password string) (*storage.User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if !s.verifyPassword(user, password) {
		return nil, nil
	}

	return &user, nil
}

func (*UserService) filterById(users []storage.User, id string) *storage.User {
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

	existingUser := s.filterById(users, user.ID)
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

func (s *UserService) EnhanceACLWithUserInfo(acl *[]storage.AccessRule) error {
	if acl != nil {
		users, err := s.storage.GetAllUsers()
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
	op storage.AccessOp,
) error {
	cfg, err := s.storage.ReadConfig()
	if err != nil {
		return err
	}

	return s.checkPermissions(cfg.ACL, userID, op, true)
}

func (s *UserService) CheckContentPermissions(
	acl *[]storage.AccessRule,
	userID string,
	op storage.AccessOp,
) error {
	return s.checkPermissions(*acl, userID, op, false)
}

func (s *UserService) checkPermissions(acl []storage.AccessRule, userID string, op storage.AccessOp, aclIsApp bool) error {

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
	if s.compareACL(acl, "user:"+userID, storage.AccessOpAdmin) {
		return nil
	}

	// Deny access
	return &AccessDeniedError{
		StatusCode: http.StatusForbidden,
	}
}

func (*UserService) compareACL(acl []storage.AccessRule, subject string, op storage.AccessOp) bool {
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
