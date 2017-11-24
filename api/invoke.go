package api

import (
	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
)

//Invoke a function
func Invoke(arg options.InvokeOptions) (*[]byte, error) {

	var services stack.Services
	var yamlGateway string

	functionName = arg.FunctionName

	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(arg.YamlFile, arg.Regex, arg.Filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress := getGatewayURL(arg.Gateway, DefaultGateway, yamlGateway)
	functionInput := arg.Input

	response, err := proxy.InvokeFunction(gatewayAddress, functionName, &functionInput, arg.ContentType, arg.Query)
	if err != nil {
		return nil, err
	}

	return response, nil
}
