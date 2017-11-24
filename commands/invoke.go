// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/api"
	"github.com/openfaas/faas-cli/options"
	"github.com/spf13/cobra"
)

var (
	contentType string
	query       []string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	invokeCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	invokeCmd.Flags().StringVarP(&gateway, "gateway", "g", deploy.DefaultGateway, "Gateway URL starting with http(s)://")

	invokeCmd.Flags().StringVar(&contentType, "content-type", "text/plain", "The content-type HTTP header such as application/json")
	invokeCmd.Flags().StringArrayVar(&query, "query", []string{}, "pass query-string options")

	faasCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   `invoke FUNCTION_NAME [--gateway GATEWAY_URL] [--content-type CONTENT_TYPE] [--query PARAM=VALUE]`,
	Short: "Invoke an OpenFaaS function",
	Long:  `Invokes an OpenFaaS function and reads from STDIN for the body of the request`,
	Example: `  faas-cli invoke echo --gateway https://domain:port
  faas-cli invoke echo --gateway https://domain:port --content-type application/json
  faas-cli invoke env --query repo=faas-cli --query org=openfaas`,
	RunE: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) error {

	functionName := ""
	if len(args) > 0 {
		functionName = args[0]
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read standard input: %s", err.Error())
	}

	response, err := api.Invoke(options.InvokeOptions{
		FaasOptions: options.FaasOptions {
			YamlFile: yamlFile,
			Regex: regex,
			Filter: filter,
		},
		SharedOptions: options.SharedOptions {
			Gateway: gateway,
		},
		ContentType: contentType,
		Query: query,
		FunctionName: functionName,
		Input: functionInput,
	})

	if response != nil {
		os.Stdout.Write(*response)
	}

	return err
}
