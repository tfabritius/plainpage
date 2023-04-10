package storage

import (
	"errors"
	"io/fs"

	"github.com/tfabritius/plainpage/model"
)

type Storage interface {
	Exists(fsPath string) bool

	ReadFile(fsPath string) ([]byte, error)
	WriteFile(fsPath string, content []byte) error
	DeleteFile(fsPath string) error

	CreateDirectory(fsPath string) error
	ReadDirectory(fsPath string) ([]fs.FileInfo, error)
	DeleteEmptyDirectory(fsPath string) error

	// GetAllUsers returns all users
	GetAllUsers() ([]model.User, error)

	// SaveAllUsers stores all users
	SaveAllUsers(users []model.User) error

	// ReadConfig returns configuration
	ReadConfig() (model.Config, error)

	// WriteConfig saves configuration
	WriteConfig(config model.Config) error
}

var ErrNotFound = errors.New("not found")
var ErrParentFolderNotFound = errors.New("parent folder not found")
var ErrPageOrFolderExistsAlready = errors.New("page or folder exists already")
var ErrFolderNotEmpty = errors.New("folder is not empty")
var ErrInvalidUsername = errors.New("invalid username")
var ErrUserExistsAlready = errors.New("user already exists")
