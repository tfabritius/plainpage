package model

import (
	"io/fs"
)

type Storage interface {
	Exists(fsPath string) bool

	ReadFile(fsPath string) ([]byte, error)
	WriteFile(fsPath string, content []byte) error
	DeleteFile(fsPath string) error

	CreateDirectory(fsPath string) error
	ReadDirectory(fsPath string) ([]fs.FileInfo, error)
	DeleteEmptyDirectory(fsPath string) error
	DeleteDirectory(fsPath string) error

	ReadConfig() (Config, error)
	WriteConfig(config Config) error
}
