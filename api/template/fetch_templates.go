// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package template

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/proxy"
)

const (
	defaultTemplateRepository = "https://github.com/openfaas/faas-cli"
	rootLanguageDirSplitCount = 3
)

const defaultWorkDir = "./"

var workDir string

//SetWorkDirectory set the work directory used to store files
func SetWorkDirectory(dir string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	workDir = path
	return nil
}

//GetWorkDirectory return the work directory used to store files
func GetWorkDirectory() string {
	if len(workDir) > 0 {
		return workDir
	}
	return defaultWorkDir
}

//GetTemplateDirectory return the template directory
func GetTemplateDirectory() string {
	return filepath.Join(GetWorkDirectory(), "template")
}

type extractAction int

const (
	shouldExtractData extractAction = iota
	newTemplateFound
	directoryAlreadyExists
	skipWritingData
)

// FetchTemplates fetch code templates from a remote URL.
func FetchTemplates(templateURL string, overwrite bool) error {

	if len(templateURL) == 0 {
		templateURL = defaultTemplateRepository
	}

	archive, err := fetchMasterZip(templateURL, GetWorkDirectory())
	if err != nil {
		removeArchive(archive)
		return err
	}

	log.Printf("Attempting to expand templates from %s\n", archive)

	preExistingLanguages, fetchedLanguages, err := expandTemplatesFromZip(archive, overwrite)
	if err != nil {
		return err
	}

	if len(preExistingLanguages) > 0 {
		log.Printf("Cannot overwrite the following %d directories: %v\n", len(preExistingLanguages), preExistingLanguages)
	}

	log.Printf("Fetched %d template(s) : %v from %s\n", len(fetchedLanguages), fetchedLanguages, templateURL)

	err = removeArchive(archive)
	if err != nil {
		return err
	}

	return nil
}

// expandTemplatesFromZip builds a list of languages that: already exist and
// could not be overwritten and // a list of languages that are newly downloaded.
func expandTemplatesFromZip(archivePath string, overwrite bool) ([]string, []string, error) {
	var existingLanguages []string
	var fetchedLanguages []string

	availableLanguages := make(map[string]bool)

	zipFile, err := zip.OpenReader(archivePath)
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	defer zipFile.Close()

	for _, z := range zipFile.File {

		relativePath := z.Name[strings.Index(z.Name, "/")+1:]
		if strings.Index(relativePath, "template/") != 0 {
			// Process only directories inside "template" at root
			continue
		}

		absolutePath := filepath.Join(GetWorkDirectory(), relativePath)

		// We know that this path is a directory if the last character is a "/"
		isDirectory := strings.HasSuffix(relativePath, "/")
		action, language := canExpandTemplateData(availableLanguages, absolutePath, overwrite, isDirectory)

		var expandFromZip bool

		switch action {

		case shouldExtractData:
			expandFromZip = true
		case newTemplateFound:
			expandFromZip = true
			fetchedLanguages = append(fetchedLanguages, language)
		case directoryAlreadyExists:
			expandFromZip = false
			existingLanguages = append(existingLanguages, language)
		case skipWritingData:
			expandFromZip = false
		default:
			return nil, nil, fmt.Errorf(fmt.Sprintf("don't know what to do when extracting zip: %s", archivePath))
		}

		if expandFromZip {
			var rc io.ReadCloser

			if rc, err = z.Open(); err != nil {
				break
			}
			defer rc.Close()

			if err = createPath(absolutePath, z.Mode()); err != nil {
				break
			}

			// If relativePath is just a directory, then skip expanding it.
			if len(absolutePath) > 1 && !isDirectory {
				if err = writeFile(rc, z.UncompressedSize64, absolutePath, z.Mode()); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	return existingLanguages, fetchedLanguages, nil
}

// canExpandTemplateData returns what we should do with the file in form of ExtractAction enum
// with the language name and whether it is a directory
func canExpandTemplateData(availableLanguages map[string]bool, absolutePath string, overwrite bool, isDirectory bool) (extractAction, string) {

	relativePath := absolutePath[len(GetWorkDirectory())+1:]

	if isDirectory {
		relativePath += "/"
	}

	if pathSplit := strings.Split(relativePath, "/"); len(pathSplit) > 2 {

		language := pathSplit[1]

		// Check if this is the root directory for a language (at ./template/lang)
		if len(pathSplit) == rootLanguageDirSplitCount && isDirectory {
			if !canWriteLanguage(availableLanguages, language, overwrite) {
				return directoryAlreadyExists, language
			}
			return newTemplateFound, language
		}

		if canWriteLanguage(availableLanguages, language, overwrite) == false {
			return skipWritingData, language
		}

		return shouldExtractData, language
	}
	// template/
	return skipWritingData, ""
}

// fetchMasterZip downloads a zip file from a repository URL
func fetchMasterZip(templateBaseURL, destinationPath string) (string, error) {
	var err error

	templateURL, err := url.Parse(templateBaseURL)
	if err != nil {
		return "", err
	}
	templateURL.Path = filepath.Join(templateURL.Path, "/archive/master.zip")

	archive := filepath.Join(destinationPath, "master.zip")

	if _, serr := os.Stat(archive); serr == nil {
		removeArchive(archive)
	}

	timeout := 120 * time.Second
	client := proxy.MakeHTTPClient(&timeout)

	req, rerr := http.NewRequest(http.MethodGet, templateURL.String(), nil)
	if rerr != nil {
		return "", rerr
	}

	log.Printf("HTTP GET %s\n", templateURL)
	res, derr := client.Do(req)
	if derr != nil {
		return "", derr
	}

	if res.StatusCode != http.StatusOK {
		ferr := fmt.Errorf(fmt.Sprintf("%s is not valid, status code %d", templateURL, res.StatusCode))
		log.Println(ferr.Error())
		return "", ferr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, rerr := ioutil.ReadAll(res.Body)
	if rerr != nil {
		return "", rerr
	}

	log.Printf("Writing %dKb to %s\n", len(bytesOut)/1024, archive)
	err = ioutil.WriteFile(archive, bytesOut, 0700)
	if err != nil {
		return "", err
	}

	return archive, err
}

// canWriteLanguage tells whether the language can be expanded from the zip or not.
// availableLanguages map keeps track of which languages we know to be okay to copy.
// overwrite flag will allow to force copy the language template
func canWriteLanguage(availableLanguages map[string]bool, language string, overwrite bool) bool {
	canWrite := false
	if availableLanguages != nil && len(language) > 0 {
		if _, found := availableLanguages[language]; found {
			return availableLanguages[language]
		}
		canWrite = templateFolderExists(language, overwrite)
		availableLanguages[language] = canWrite
	}

	return canWrite
}
