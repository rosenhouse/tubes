package application

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FilesystemConfigStore struct {
	RootDir string
}

func (s *FilesystemConfigStore) getFilePath(key string) (string, error) {
	rootDir, err := filepath.Abs(s.RootDir)
	if err != nil {
		return "", err // not tested
	}
	fileInfo, err := os.Stat(rootDir)
	if err != nil {
		return "", err
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("expecting %s to be a directory", rootDir)
	}
	return filepath.Join(rootDir, filepath.FromSlash(key)), nil
}

func (s *FilesystemConfigStore) Get(key string) ([]byte, error) {
	filePath, err := s.getFilePath(key)
	if err != nil {
		return nil, err
	}

	value, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return value, nil
}

const (
	dirCreationBits  = 0777
	fileCreationBits = 0600
)

func (s *FilesystemConfigStore) Set(key string, value []byte) error {
	filePath, err := s.getFilePath(key)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(filePath), dirCreationBits)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, value, fileCreationBits)
	if err != nil {
		return err
	}
	return nil
}

func (s *FilesystemConfigStore) IsEmpty() (bool, error) {
	fs, err := ioutil.ReadDir(s.RootDir)
	if err != nil {
		return false, err
	}

	return len(fs) == 0, nil
}
