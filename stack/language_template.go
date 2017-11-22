// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func ParseYAMLForLanguageTemplate(file string) (*LanguageTemplate, error) {
	var err error
	var fileData []byte

	urlParsed, err := url.Parse(file)
	if err == nil && len(urlParsed.Scheme) > 0 {
		fmt.Println("Parsed: " + urlParsed.String())
		fileData, err = fetchYAML(urlParsed)
		if err != nil {
			return nil, err
		}
	} else {
		fileData, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
	}

	return ParseYAMLDataForLanguageTemplate(fileData)
}

// ParseYAMLDataForLanguageTemplate parses YAML data into language template
func ParseYAMLDataForLanguageTemplate(fileData []byte) (*LanguageTemplate, error) {
	var langTemplate LanguageTemplate
	var err error

	err = yaml.Unmarshal(fileData, &langTemplate)
	if err != nil {
		fmt.Printf("Error with YAML file\n")
		return nil, err
	}

	return &langTemplate, err
}

func IsValidTemplate(lang string) bool {
	templatePath := filepath.Join(
		os.Getenv("workdir"),
		"./template",
		lang,
	)
	var found bool
	if strings.ToLower(lang) == "dockerfile" {
		found = true
	} else if _, err := os.Stat(templatePath); err == nil {
		templateYAMLPath := templatePath + "/template.yml"

		_, err := ParseYAMLForLanguageTemplate(templateYAMLPath)
		if err == nil {
			found = true
		}
	}

	return found
}
