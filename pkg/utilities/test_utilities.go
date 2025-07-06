package utilities

import (
	"log"
	"os"
)

func CreateTempFile(dir string, pattern string, fileContent []byte) (*os.File, error) {
	// Create a temp YAML file
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil {
			log.Printf("Error removing file: %v", removeErr)
		}
	}()
	if _, err := file.Write(fileContent); err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Error closing file: %v", closeErr)
		}
	}()
	return file, nil
}
