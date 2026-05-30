package persistence

import (
	"os"
	"path/filepath"
)

func InitCurseDir(root string) error {
	dirs := []string{
		filepath.Join(root, "logs"),
		filepath.Join(root, "staging"),
		filepath.Join(root, "missions"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func DefaultRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".curse"), nil
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
