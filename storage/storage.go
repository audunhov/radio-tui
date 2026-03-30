package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"radio-tui/models"
)

type StorageService struct {
	filePath string
}

func NewStorageService() (*StorageService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(configDir, "radio-tui")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &StorageService{
		filePath: filepath.Join(dir, "favorites.json"),
	}, nil
}

func (s *StorageService) LoadFavorites() ([]models.Station, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []models.Station{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var favs []models.Station
	if err := json.Unmarshal(data, &favs); err != nil {
		return nil, err
	}

	return favs, nil
}

func (s *StorageService) SaveFavorites(favs []models.Station) error {
	data, err := json.MarshalIndent(favs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}
