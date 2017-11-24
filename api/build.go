// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package api

import (
	"fmt"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/stack"
)

//Build one or more functions
func Build(arg options.BuildOptions) error {

	var services stack.Services
	if arg.Services != nil {
		services = *arg.Services
	} else {
		if len(arg.YamlFile) > 0 {
			parsedServices, err := stack.ParseYAMLFile(arg.YamlFile, arg.Regex, arg.Filter)
			if err != nil {
				return err
			}

			if parsedServices != nil {
				services = *parsedServices
			}
		}
	}

	if pullErr := Pull(options.TemplatePullOptions{URL: ""}); pullErr != nil {
		return fmt.Errorf("could not pull templates for OpenFaaS: %v", pullErr)
	}

	if len(services.Functions) > 0 {
		return builder.BuildStack(&services, arg.Parallel, arg.Nocache, arg.Squash, arg.Shrinkwrap)
	}

	if len(arg.Image) == 0 {
		return fmt.Errorf("please provide a valid --image name for your Docker image")
	}
	if len(arg.Handler) == 0 {
		return fmt.Errorf("please provide the full path to your function's handler")
	}
	if len(arg.FunctionName) == 0 {
		return fmt.Errorf("please provide the deployed --name of your function")
	}
	if err := builder.BuildImage(
		arg.Image,
		arg.Handler,
		arg.FunctionName,
		arg.Language,
		arg.Nocache,
		arg.Squash,
		arg.Shrinkwrap,
	); err != nil {
		return err
	}

	return nil
}
