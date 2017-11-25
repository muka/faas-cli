package api

import (
	"log"

	"github.com/openfaas/faas-cli/api/template"
	"github.com/openfaas/faas-cli/options"
)

//Pull templates from an URL
func Pull(arg options.TemplatePullOptions) error {
	repository := arg.URL
	log.Printf("Fetch templates from repository: %s\n", repository)
	err := template.FetchTemplates(arg.URL, arg.Overwrite)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	return err
}
