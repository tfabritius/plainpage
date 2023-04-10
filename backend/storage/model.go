package storage

import (
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

	ReadConfig() (model.Config, error)
	WriteConfig(config model.Config) error
}
