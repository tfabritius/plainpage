package service

import (
	"io/fs"
	"testing"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

type mockStorage struct {
	users []byte
}

func (m *mockStorage) Exists(path string) bool {
	if path == "users.yml" {
		// return len(m.users) > 0
		return m.users != nil
	}
	panic("not supported")
}

func (m *mockStorage) ReadFile(path string) ([]byte, error) {
	if path == "users.yml" {
		return m.users, nil
	}
	panic("not supported")
}

func (m *mockStorage) WriteFile(path string, data []byte) error {
	if path == "users.yml" {
		m.users = data
		return nil
	}
	panic("not supported")
}

func (m *mockStorage) DeleteFile(path string) error {
	panic("not supported")
}

func (m *mockStorage) CreateDirectory(path string) error {
	panic("not supported")
}

func (m *mockStorage) ReadConfig() (model.Config, error) {
	panic("not supported")
}

func (m *mockStorage) WriteConfig(config model.Config) error {
	panic("not supported")
}

func (m *mockStorage) ReadDirectory(fsPath string) ([]fs.FileInfo, error) {
	panic("not supported")
}

func (m *mockStorage) DeleteEmptyDirectory(fsPath string) error {
	panic("not supported")
}

func (m *mockStorage) DeleteDirectory(fsPath string) error {
	panic("not supported")
}

func TestUserService_Create(t *testing.T) {
	r := require.New(t)
	mock := &mockStorage{}
	userService := NewUserService(mock)

	username := "testuser"
	password := "testpassword"
	displayName := "Test User"

	user, err := userService.Create(username, password, displayName)
	r.NoError(err)
	r.NotNil(user)
	r.Equal(displayName, user.DisplayName)
	r.Equal(username, user.Username)

	// Test non-unique username
	duplicateUsername := "TestUser"
	_, err = userService.Create(duplicateUsername, password, displayName)
	r.Error(err)
	r.ErrorIs(err, model.ErrUserExistsAlready)

	// Test invalid username
	invalidUsername := "user!"
	_, err = userService.Create(invalidUsername, password, displayName)
	r.Error(err)
	r.ErrorIs(err, model.ErrInvalidUsername)
}

func TestUserService_GetByUsername(t *testing.T) {
	r := require.New(t)
	mock := &mockStorage{}
	userService := NewUserService(mock)

	var cUserId string
	for _, i := range []string{"a", "b", "c", "d", "e"} {
		user, err := userService.Create("testuser-"+i, "test-password", "Test Name "+i)
		r.NoError(err)
		if i == "c" {
			cUserId = user.ID
		}
	}

	foundUser, err := userService.GetByUsername("testuser-c")
	r.NoError(err)
	r.Equal(cUserId, foundUser.ID)

	_, err = userService.GetByUsername("testuser-x")
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)
}

func TestUserService_GetById(t *testing.T) {
	r := require.New(t)
	mock := &mockStorage{}
	userService := NewUserService(mock)

	var cUserId string
	for _, i := range []string{"a", "b", "c", "d", "e"} {
		user, err := userService.Create("testuser-"+i, "test-password", "Test Name "+i)
		r.NoError(err)
		if i == "c" {
			cUserId = user.ID
		}
	}

	foundUser, err := userService.GetById(cUserId)
	r.NoError(err)
	r.Equal("testuser-c", foundUser.Username)

	_, err = userService.GetById("xxx")
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)
}

func TestUserService_VerifyCredentials(t *testing.T) {
	r := require.New(t)
	mock := &mockStorage{}
	userService := NewUserService(mock)

	username := "testuser"
	password := "testpassword"
	displayName := "Test User"

	user, err := userService.Create(username, password, displayName)
	r.NoError(err)

	foundUser, err := userService.VerifyCredentials(username, password)
	r.NoError(err)
	r.NotNil(foundUser)
	r.Equal(user, *foundUser)

	// Test unknown hash algorithm
	user.PasswordHash = "xxx:" + password
	err = userService.Save(user)
	r.NoError(err)

	foundUser, err = userService.VerifyCredentials(username, password)
	r.NoError(err)
	r.Nil(foundUser)

	// Test plain passwords
	user.PasswordHash = "plain:" + password
	err = userService.Save(user)
	r.NoError(err)

	foundUser, err = userService.VerifyCredentials(username, password)
	r.NoError(err)
	r.NotNil(foundUser)
	// On login password should be hashed automatically
	r.Equal(user.ID, foundUser.ID)
	r.Equal(user.Username, foundUser.Username)
	r.Equal(user.DisplayName, foundUser.DisplayName)
	r.False(strings.HasPrefix(foundUser.PasswordHash, "plain:"))

	// After successful login with plain password, it should be hashed and persisted
	fetchedUser, err := userService.GetByUsername(username)
	r.NoError(err)
	r.False(strings.HasPrefix(fetchedUser.PasswordHash, "plain:"))

	// Test invalid password
	foundUser, err = userService.VerifyCredentials(username, "wrongpassword")
	r.NoError(err)
	r.Nil(foundUser)

	// Test not found user
	foundUser, err = userService.VerifyCredentials("wronguser", password)
	r.NoError(err)
	r.Nil(foundUser)
}

func TestUserService_Save(t *testing.T) {
	r := require.New(t)

	mock := &mockStorage{}
	userService := NewUserService(mock)

	username := "testuser"
	password := "testpassword"
	displayName := "Test User"

	user, err := userService.Create(username, password, displayName)
	r.NoError(err)

	newDisplayName := "New Test User"
	newUsername := "new-testuser"
	user.DisplayName = newDisplayName
	user.Username = newUsername

	err = userService.Save(user)
	r.NoError(err)

	user, err = userService.GetById(user.ID)
	r.NoError(err)
	r.Equal(newDisplayName, user.DisplayName)
	r.Equal(newUsername, user.Username)

	// Test non-unique username
	otherUsername := "otheruser"
	_, err = userService.Create(otherUsername, password, displayName)
	r.NoError(err)
	user.Username = otherUsername
	err = userService.Save(user)
	r.Error(err)
	r.ErrorIs(err, model.ErrUserExistsAlready)

	// Test unknown ID
	user.ID = "xxx"
	err = userService.Save(user)
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)
}

func TestUserService_DeleteByUsername(t *testing.T) {
	r := require.New(t)

	mock := &mockStorage{}
	userService := NewUserService(mock)

	for _, i := range []string{"a", "b", "c", "d", "e"} {
		_, err := userService.Create("testuser-"+i, "test-password", "Test Name "+i)
		r.NoError(err)
	}

	err := userService.DeleteByUsername("testuser-c")
	r.NoError(err)

	_, err = userService.GetByUsername("testuser-c")
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)

	err = userService.DeleteByUsername("testuser-x")
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)
}
