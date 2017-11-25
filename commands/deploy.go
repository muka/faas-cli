// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/openfaas/faas-cli/api"
	"github.com/openfaas/faas-cli/options"
	"github.com/spf13/cobra"
)

// Flags that are to be added to commands.
var (
	envvarOpts  []string
	replace     bool
	update      bool
	constraints []string
	secrets     []string
	labelOpts   []string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	deployCmd.Flags().StringVar(&fprocess, "fprocess", "", "Fprocess to be run by the watchdog")
	deployCmd.Flags().StringVarP(&gateway, "gateway", "g", api.DefaultGateway, "Gateway URL starting with http(s)://")
	deployCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	deployCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	deployCmd.Flags().StringVar(&language, "lang", "", "Programming language template")
	deployCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	deployCmd.Flags().StringVar(&network, "network", api.DefaultNetwork, "Name of the network")

	// Setup flags that are used only by this command (variables defined above)
	deployCmd.Flags().StringArrayVarP(&envvarOpts, "env", "e", []string{}, "Set one or more environment variables (ENVVAR=VALUE)")

	deployCmd.Flags().StringArrayVarP(&labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")

	deployCmd.Flags().BoolVar(&replace, "replace", true, "Replace any existing function")
	deployCmd.Flags().BoolVar(&update, "update", false, "Update existing functions")

	deployCmd.Flags().StringArrayVar(&constraints, "constraint", []string{}, "Apply a constraint to the function")
	deployCmd.Flags().StringArrayVar(&secrets, "secret", []string{}, "Give the function access to a secure secret")

	// Set bash-completion.
	_ = deployCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	faasCmd.AddCommand(deployCmd)
}

// deployCmd handles deploying OpenFaaS function containers
var deployCmd = &cobra.Command{
	Use: `deploy -f YAML_FILE [--replace=false]
  faas-cli deploy --image IMAGE_NAME
                  --name FUNCTION_NAME
                  [--lang <ruby|python|node|csharp>]
                  [--gateway GATEWAY_URL]
                  [--network NETWORK_NAME]
                  [--handler HANDLER_DIR]
                  [--fprocess PROCESS]
                  [--env ENVVAR=VALUE ...]
                  [--label LABEL=VALUE ...]
				  [--replace=false]
				  [--update=false]
                  [--constraint PLACEMENT_CONSTRAINT ...]
                  [--regex "REGEX"]
                  [--filter "WILDCARD"]
				  [--secret "SECRET_NAME"]`,

	Short: "Deploy OpenFaaS functions",
	Long: `Deploys OpenFaaS function containers either via the supplied YAML config using
the "--yaml" flag (which may contain multiple function definitions), or directly
via flags. Note: --replace and --update are mutually exclusive.`,
	Example: `  faas-cli deploy -f https://domain/path/myfunctions.yml
  faas-cli deploy -f ./samples.yml
  faas-cli deploy -f ./samples.yml --label canary=true
  faas-cli deploy -f ./samples.yml --filter "*gif*" --secret dockerhuborg
  faas-cli deploy -f ./samples.yml --regex "fn[0-9]_.*"
  faas-cli deploy -f ./samples.yml --replace=false
  faas-cli deploy -f ./samples.yml --update=true
  faas-cli deploy --image=alexellis/faas-url-ping --name=url-ping
  faas-cli deploy --image=my_image --name=my_fn --handler=/path/to/fn/
                  --gateway=http://remote-site.com:8080 --lang=python
                  --env=MYVAR=myval`,
	RunE: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	dargs := options.DeployOptions{
		FaasOptions: getFaasOptions(),
		SharedOptions: getSharedOptions(),
		EnvvarOpts:  envvarOpts,
		Replace:     replace,
		Update:      update,
		Constraints: constraints,
		Secrets:     secrets,
		LabelOpts:   labelOpts,
	}
	return api.Deploy(dargs)
}
