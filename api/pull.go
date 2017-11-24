package api

import (
	"fmt"

	"github.com/openfaas/faas-cli/api/template"
	"github.com/openfaas/faas-cli/options"
)

//Pull templates from an URL
func Pull(arg options.TemplatePullOptions) error {
	repository := arg.URL
	fmt.Println("Fetch templates from repository: " + repository)
	err := template.FetchTemplates(arg.URL, arg.Overwrite)
	return err
}
