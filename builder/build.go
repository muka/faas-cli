// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-cli/stack"
)

// BuildImage construct Docker image from function parameters
func BuildImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool) error {

	if stack.IsValidTemplate(language) {

		var tempPath string
		if strings.ToLower(language) == "dockerfile" {

			if shrinkwrap {
				fmt.Printf("Nothing to do for: %s.\n", functionName)

				return nil
			}

			tempPath = handler
			if _, err := os.Stat(handler); err != nil {
				fmt.Printf("Unable to build %s, %s is an invalid path\n", image, handler)
				fmt.Printf("Image: %s not built.\n", image)

				return nil
			}
			fmt.Printf("Building: %s with Dockerfile. Please wait..\n", image)

		} else {

			tempPath = createBuildTemplate(functionName, handler, language)
			fmt.Printf("Building: %s with %s template. Please wait..\n", image, language)

			if shrinkwrap {
				fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)

				return nil
			}
		}

		flagStr := buildFlagString(nocache, squash, os.Getenv("http_proxy"), os.Getenv("https_proxy"))
		builder := strings.Split(fmt.Sprintf("docker build %s-t %s .", flagStr, image), " ")
		ExecCommand(tempPath, builder)
		fmt.Printf("Image: %s built.\n", image)

	} else {
		return fmt.Errorf("Language template: %s not supported. Build a custom Dockerfile instead", language)
	}

	return nil
}

// createBuildTemplate creates temporary build folder to perform a Docker build with language template
func createBuildTemplate(functionName string, handler string, language string) string {
	tempPath := filepath.Join(
		os.Getenv("workdir"),
		"./build",
		functionName,
	)

	fmt.Printf("Clearing temporary build folder: %s\n", tempPath)

	clearErr := os.RemoveAll(tempPath)
	if clearErr != nil {
		fmt.Printf("Error clearing temporary build folder %s\n", tempPath)
	}

	functionPath := filepath.Join(tempPath, "/function")

	fmt.Printf("Preparing %s %s\n", handler+"/", functionPath)

	mkdirErr := os.MkdirAll(functionPath, 0700)
	if mkdirErr != nil {
		fmt.Printf("Error creating path %s - %s.\n", functionPath, mkdirErr.Error())
	}

	// Drop in directory tree from template
	CopyFiles(filepath.Join(os.Getenv("workdir"), "./template", language), tempPath, true)

	// Overlay in user-function
	CopyFiles(handler, functionPath, true)

	return tempPath
}

// CopyFiles copies files from src to destination, optionally recursively.
func CopyFiles(src string, destination string, recursive bool) {

	files, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		if file.IsDir() == false {
			cp(filepath.Join(src, file.Name()), filepath.Join(destination, file.Name()))
		} else {
			//make new destination dir
			newDir := filepath.Join(destination, file.Name())

			if !pathExists(newDir) {

				debugPrint(fmt.Sprintf("Creating directory: %s at %s", file.Name(), newDir))
				newDirErr := os.Mkdir(newDir, 0700)

				if newDirErr != nil {
					fmt.Printf("Error creating path %s - %s.\n", newDir, newDirErr.Error())
				}
			}

			//did the call ask to recurse into sub directories?
			if recursive == true {
				//call CopyFiles to copy the contents
				CopyFiles(filepath.Join(src, file.Name()), newDir, true)
			}
		}
	}
}

func pathExists(path string) bool {
	exists := true

	if _, err := os.Stat(path); os.IsNotExist(err) {
		exists = false
	}

	return exists
}

func cp(src string, destination string) error {

	debugPrint(fmt.Sprintf("cp - %s %s", src, destination))

	memoryBuffer, readErr := ioutil.ReadFile(src)
	if readErr != nil {
		return fmt.Errorf("Error reading source file: %s\n" + readErr.Error())
	}
	writeErr := ioutil.WriteFile(destination, memoryBuffer, 0660)
	if writeErr != nil {
		return fmt.Errorf("Error writing file: %s\n" + writeErr.Error())
	}

	return nil
}

func buildFlagString(nocache bool, squash bool, httpProxy string, httpsProxy string) string {

	buildFlags := ""

	if nocache {
		buildFlags += "--no-cache "
	}
	if squash {
		buildFlags += "--squash "
	}

	if len(httpProxy) > 0 {
		buildFlags += fmt.Sprintf("--build-arg http_proxy=%s ", httpProxy)
	}

	if len(httpsProxy) > 0 {
		buildFlags += fmt.Sprintf("--build-arg https_proxy=%s ", httpsProxy)
	}

	return buildFlags
}

func debugPrint(message string) {

	if val, exists := os.LookupEnv("debug"); exists && (val == "1" || val == "true") {
		fmt.Println(message)
	}
}
