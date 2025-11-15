package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type Storage struct {
	basePath string
}

func New(basePath string) (*Storage, error) {
	storage := &Storage{basePath: basePath}
	if err := storage.ensureBaseDir(); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return storage, nil
}

func (s *Storage) ensureBaseDir() error {
	return os.MkdirAll(s.basePath, 0755)
}

func (s *Storage) GetFilePath(fileID, fileName string) string {
	return filepath.Join(s.basePath, fileID, fileName)
}

func (s *Storage) EnsureFileDir(fileID string) error {
	dir := filepath.Join(s.basePath, fileID)
	return os.MkdirAll(dir, 0755)
}

func (s *Storage) SaveFile(fileID, fileName string, data []byte) error {
	if err := s.EnsureFileDir(fileID); err != nil {
		return fmt.Errorf("failed to create file directory: %w", err)
	}

	filePath := s.GetFilePath(fileID, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Set secure permissions (600)
	if err := os.Chmod(filePath, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *Storage) FileExists(fileID, fileName string) bool {
	filePath := s.GetFilePath(fileID, fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

func (s *Storage) GetFileSize(fileID, fileName string) (int64, error) {
	filePath := s.GetFilePath(fileID, fileName)
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *Storage) DeleteFile(fileID, fileName string) error {
	filePath := s.GetFilePath(fileID, fileName)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove the directory if empty
	dir := filepath.Join(s.basePath, fileID)
	os.Remove(dir) // Ignore error if directory not empty

	return nil
}

func (s *Storage) DeleteFileDir(fileID string) error {
	dir := filepath.Join(s.basePath, fileID)
	return os.RemoveAll(dir)
}

