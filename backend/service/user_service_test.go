package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

func TestUserService_Create(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
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
	mock := newMockStorage()
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
	mock := newMockStorage()
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
	mock := newMockStorage()
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
	mock := newMockStorage()
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
	mock := newMockStorage()
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
