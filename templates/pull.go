package templates

import (
	"fmt"

	"github.com/openfaas/faas-cli/options"
)

//Pull templates from an URL
func Pull(arg options.TemplatePullOptions) error {
	repository := arg.URL
	fmt.Println("Fetch templates from repository: " + repository)
	err := fetchTemplates(arg.URL, arg.Overwrite)
	return err
}

//PullTemplates fetch templates from an URL
func PullTemplates(repository string) error {
	return Pull(options.TemplatePullOptions{
		URL: repository,
	})
}
