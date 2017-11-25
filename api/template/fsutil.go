package template

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func writeFile(rc io.ReadCloser, size uint64, relativePath string, perms os.FileMode) error {
	var err error

	defer rc.Close()
	f, err := os.OpenFile(relativePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.CopyN(f, rc, int64(size))

	return err
}

func createPath(relativePath string, perms os.FileMode) error {
	dir := filepath.Dir(relativePath)
	//HACK forcing 0755 to avoid weird permission errors
	err := os.MkdirAll(dir, 0755)
	return err
}

// Takes a language input (e.g. "node"), tells whether or not it is OK to download
func templateFolderExists(language string, overwrite bool) bool {
	dir := filepath.Join(GetTemplateDirectory(), language)
	if _, err := os.Stat(dir); err == nil && !overwrite {
		// The directory template/language/ exists
		return false
	}
	return true
}

// removeArchive removes the given file
func removeArchive(archive string) error {
	log.Printf("Cleaning up zip file...")
	var err error

	if _, err = os.Stat(archive); err == nil {
		err = os.Remove(archive)
	}

	return err
}
