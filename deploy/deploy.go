// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
)

const (
	DefaultGateway = "http://localhost:8080"
	DefaultNetwork = "func_functions"
	DefaultYAML    = "stack.yml"
)

//Deploy a function
func Deploy(arg options.DeployOptions) error {

	if arg.Update && arg.Replace {
		fmt.Println(`Cannot specify --update and --replace at the same time.
  --replace    removes an existing deployment before re-creating it
  --update     provides a rolling update to a new function image or configuration`)
		return fmt.Errorf("cannot specify --update and --replace at the same time")
	}

	var services stack.Services
	if arg.Services != nil {
		services = *arg.Services
	} else {
		if len(arg.YamlFile) > 0 {
			parsedServices, err := stack.ParseYAMLFile(arg.YamlFile, arg.Regex, arg.Filter)
			if err != nil {
				return err
			}

			parsedServices.Provider.GatewayURL = getGatewayURL(arg.Gateway, DefaultGateway, parsedServices.Provider.GatewayURL)

			// Override network if passed
			if len(arg.Network) > 0 && arg.Network != DefaultNetwork {
				parsedServices.Provider.Network = arg.Network
			}

			if parsedServices != nil {
				services = *parsedServices
			}
		}
	}

	if len(services.Functions) > 0 {
		if len(services.Provider.Network) == 0 {
			services.Provider.Network = DefaultNetwork
		}

		for k, function := range services.Functions {

			function.Name = k
			if arg.Update {
				fmt.Printf("Updating: %s.\n", function.Name)
			} else {
				fmt.Printf("Deploying: %s.\n", function.Name)
			}

			var functionConstraints []string
			if function.Constraints != nil {
				functionConstraints = *function.Constraints
			} else if len(arg.Constraints) > 0 {
				functionConstraints = arg.Constraints
			}

			fileEnvironment, err := readFiles(function.EnvironmentFile)
			if err != nil {
				return err
			}

			labelMap := map[string]string{}
			if function.Labels != nil {
				labelMap = *function.Labels
			}

			labelArgumentMap, labelErr := parseMap(arg.LabelOpts, "label")
			if labelErr != nil {
				return fmt.Errorf("error parsing labels: %v", labelErr)
			}

			allLabels := mergeMap(labelMap, labelArgumentMap)

			allEnvironment, envErr := compileEnvironment(arg.EnvvarOpts, function.Environment, fileEnvironment)
			if envErr != nil {
				return envErr
			}

			// Get FProcess to use from the ./template/template.yml, if a template is being used
			if languageExistsNotDockerfile(function.Language) {
				var fprocessErr error
				function.FProcess, fprocessErr = deriveFprocess(function)
				if fprocessErr != nil {
					return fprocessErr
				}
			}

			functionResourceRequest1 := proxy.FunctionResourceRequest{
				Limits:   function.Limits,
				Requests: function.Requests,
			}

			proxy.DeployFunction(
				function.FProcess,
				services.Provider.GatewayURL,
				function.Name,
				function.Image,
				function.Language,
				arg.Replace,
				allEnvironment,
				services.Provider.Network,
				functionConstraints,
				arg.Update,
				arg.Secrets,
				allLabels,
				functionResourceRequest1,
			)
		}
	} else {
		if len(arg.Image) == 0 {
			return fmt.Errorf("please provide a --image to be deployed")
		}
		if len(arg.FunctionName) == 0 {
			return fmt.Errorf("please provide a --name for your function as it will be deployed on FaaS")
		}

		envvars, err := parseMap(arg.EnvvarOpts, "env")
		if err != nil {
			return fmt.Errorf("error parsing envvars: %v", err)
		}

		labelMap, labelErr := parseMap(arg.LabelOpts, "label")
		if labelErr != nil {
			return fmt.Errorf("error parsing labels: %v", labelErr)
		}
		functionResourceRequest1 := proxy.FunctionResourceRequest{}
		proxy.DeployFunction(
			arg.Fprocess,
			arg.Gateway,
			arg.FunctionName,
			arg.Image,
			arg.Language,
			arg.Replace,
			envvars,
			arg.Network,
			arg.Constraints,
			arg.Update,
			arg.Secrets,
			labelMap,
			functionResourceRequest1,
		)
	}

	return nil
}

func buildLabelMap(labelOpts []string) map[string]string {
	labelMap := map[string]string{}
	for _, opt := range labelOpts {
		if !strings.Contains(opt, "=") {
			fmt.Println("Error - label option does not contain a value")
		} else {
			index := strings.Index(opt, "=")

			labelMap[opt[0:index]] = opt[index+1:]
		}
	}
	return labelMap
}

func readFiles(files []string) (map[string]string, error) {
	envs := make(map[string]string)

	for _, file := range files {
		bytesOut, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return nil, readErr
		}

		envFile := stack.EnvironmentFile{}
		unmarshalErr := yaml.Unmarshal(bytesOut, &envFile)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		for k, v := range envFile.Environment {
			envs[k] = v
		}
	}
	return envs, nil
}

func parseMap(envvars []string, keyName string) (map[string]string, error) {
	result := make(map[string]string)
	for _, envvar := range envvars {
		s := strings.SplitN(strings.TrimSpace(envvar), "=", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("label format is not correct, needs key=value")
		}
		envvarName := s[0]
		envvarValue := s[1]

		if !(len(envvarName) > 0) {
			return nil, fmt.Errorf("Empty %s name: [%s]", keyName, envvar)
		}
		if !(len(envvarValue) > 0) {
			return nil, fmt.Errorf("Empty %s value: [%s]", keyName, envvar)
		}

		result[envvarName] = envvarValue
	}
	return result, nil
}

func mergeMap(i map[string]string, j map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range i {
		merged[k] = v
	}
	for k, v := range j {
		merged[k] = v
	}
	return merged
}

func getGatewayURL(argumentURL string, defaultURL string, yamlURL string) string {
	var gatewayURL string

	if len(argumentURL) > 0 && argumentURL != defaultURL {
		gatewayURL = argumentURL
	} else if len(yamlURL) > 0 {
		gatewayURL = yamlURL
	} else {
		gatewayURL = defaultURL
	}

	return gatewayURL
}

func compileEnvironment(envvarOpts []string, yamlEnvironment map[string]string, fileEnvironment map[string]string) (map[string]string, error) {
	envvarArguments, err := parseMap(envvarOpts, "env")
	if err != nil {
		return nil, fmt.Errorf("error parsing envvars: %v", err)
	}

	functionAndStack := mergeMap(yamlEnvironment, fileEnvironment)
	return mergeMap(functionAndStack, envvarArguments), nil
}

func deriveFprocess(function stack.Function) (string, error) {
	var fprocess string

	pathToTemplateYAML := filepath.Join(os.Getenv("workdir"), "template", function.Language, "template.yml")
	if _, err := os.Stat(pathToTemplateYAML); os.IsNotExist(err) {
		return "", err
	}

	var langTemplate stack.LanguageTemplate
	parsedLangTemplate, err := stack.ParseYAMLForLanguageTemplate(pathToTemplateYAML)

	if err != nil {
		return "", err

	}

	if parsedLangTemplate != nil {
		langTemplate = *parsedLangTemplate
		fprocess = langTemplate.FProcess
	}

	return fprocess, nil
}

func languageExistsNotDockerfile(language string) bool {
	return len(language) > 0 && strings.ToLower(language) != "dockerfile"
}
