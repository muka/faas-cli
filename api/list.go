package api

import (
	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas/gateway/requests"
)

//List available functions
func List(arg options.ListOptions) ([]requests.Function, error) {

	var services stack.Services
	var gatewayAddress string
	var yamlGateway string

	if arg.Services != nil {
		services = *arg.Services
	} else {
		if len(arg.YamlFile) > 0 {
			parsedServices, err := stack.ParseYAMLFile(arg.YamlFile, arg.Regex, arg.Filter)
			if err != nil {
				return nil, err
			}

			if parsedServices != nil {
				services = *parsedServices
				yamlGateway = services.Provider.GatewayURL
			}
		}
	}

	gatewayAddress = GetGatewayURL(arg.Gateway, DefaultGateway, yamlGateway)

	functions, err := proxy.ListFunctions(gatewayAddress)
	if err != nil {
		return nil, err
	}

	return functions, nil
}
