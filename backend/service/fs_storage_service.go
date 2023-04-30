package service

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/tfabritius/plainpage/model"
	"gopkg.in/yaml.v3"
)

type fsStorage struct {
	DataDir string
}

func NewFsStorage(dataDir string) model.Storage {
	log.Println("Data directory:", dataDir)

	fi, err := os.Stat(dataDir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(dataDir, 0700); err != nil {
			log.Fatalln("Could not create data directory:", err)
		}
		log.Println("Data directory created")
	} else if err != nil {
		log.Fatalln("Cannot access data directory:", err)
	} else if !fi.IsDir() {
		log.Fatalln("Data directory is not a directory")
	}

	storage := fsStorage{DataDir: dataDir}

	return &storage
}

func (fss *fsStorage) Exists(fsPath string) bool {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	_, err := os.Stat(fsPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (fss *fsStorage) ReadFile(fsPath string) ([]byte, error) {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	bytes, err := os.ReadFile(fsPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	return bytes, nil
}

func (fss *fsStorage) WriteFile(fsPath string, content []byte) error {
	fsPath = filepath.Join(fss.DataDir, fsPath)

	if err := fss.createDir(fsPath); err != nil {
		return fmt.Errorf("could not createDir: %w", err)
	}

	if err := os.WriteFile(fsPath, content, 0600); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

func (fss *fsStorage) DeleteFile(fsPath string) error {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	err := os.Remove(fsPath)
	if err != nil {
		return fmt.Errorf("could not remove file: %w", err)
	}
	return nil
}

func (fss *fsStorage) CreateDirectory(fsPath string) error {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	if err := os.Mkdir(fsPath, 0700); err != nil {
		return err
	}
	return nil
}

func (fss *fsStorage) ReadDirectory(fsPath string) ([]fs.FileInfo, error) {
	dirPath := filepath.Join(fss.DataDir, fsPath)

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

	return fileInfos, nil
}

func (fss *fsStorage) DeleteEmptyDirectory(fsPath string) error {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	return os.Remove(fsPath)
}

func (fss *fsStorage) DeleteDirectory(fsPath string) error {
	fsPath = filepath.Join(fss.DataDir, fsPath)
	return os.RemoveAll(fsPath)
}

func (fss *fsStorage) createDir(file string) error {
	dir := filepath.Dir(file)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("could not create directories: %w", err)
	}

	return nil
}

func (fss *fsStorage) ReadConfig() (model.Config, error) {
	bytes, err := fss.ReadFile("config.yml")
	if err != nil {
		return model.Config{}, fmt.Errorf("could not read config.yml: %w", err)
	}

	// parse YAML
	config := model.Config{}
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return model.Config{}, fmt.Errorf("could not parse YAML: %w", err)
	}

	return config, nil
}

func (fss *fsStorage) WriteConfig(config model.Config) error {
	bytes, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := fss.WriteFile("config.yml", bytes); err != nil {
		return fmt.Errorf("could not write config.yml: %w", err)
	}

	return nil
}
