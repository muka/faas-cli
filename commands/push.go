// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/api"
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(pushCmd)
	pushCmd.Flags().IntVar(&parallel, "parallel", 1, "Push images in parallel to depth specified.")
}

// pushCmd handles pushing function container images to a remote repo
var pushCmd = &cobra.Command{
	Use:   `push -f YAML_FILE [--regex "REGEX"] [--filter "WILDCARD"] [--parallel]`,
	Short: "Push OpenFaaS functions to remote registry (Docker Hub)",
	Long: `Pushes the OpenFaaS function container image(s) defined in the supplied YAML
config to a remote repository.

These container images must already be present in your local image cache.`,

	Example: `  faas-cli push -f https://domain/path/myfunctions.yml
  faas-cli push -f ./samples.yml
  faas-cli push -f ./samples.yml --parallel 4
  faas-cli push -f ./samples.yml --filter "*gif*"
  faas-cli push -f ./samples.yml --regex "fn[0-9]_.*"`,
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	return api.Push(options.PushOptions{
		FaasOptions: options.FaasOptions{
			YamlFile: yamlFile,
			Regex: regex,
			Filter: filter,
		},
		SharedOptions: options.SharedOptions{
			Network: network,
			Image: image,
			Handler: handler,
			FunctionName: functionName,
			Language: language,
		},
		Parallel: parallel,
	})
}
