package storage

import (
	"os"
	"path/filepath"
	"radio-tui/models"
	"testing"
)

func TestStorageService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "radio-tui-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s := &StorageService{
		filePath: filepath.Join(tempDir, "favorites.json"),
	}

	// Load empty
	favs, err := s.LoadFavorites()
	if err != nil {
		t.Fatalf("LoadFavorites failed: %v", err)
	}
	if len(favs) != 0 {
		t.Errorf("Expected 0 favorites, got %d", len(favs))
	}

	// Save and load
	testFavs := []models.Station{
		{ID: "1", Name: "Station 1", URL: "http://test.com"},
	}
	err = s.SaveFavorites(testFavs)
	if err != nil {
		t.Fatalf("SaveFavorites failed: %v", err)
	}

	favs, err = s.LoadFavorites()
	if err != nil {
		t.Fatalf("LoadFavorites failed: %v", err)
	}
	if len(favs) != 1 || favs[0].Name != "Station 1" {
		t.Errorf("Expected 1 favorite named 'Station 1', got %+v", favs)
	}
}
