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
	nocache    bool
	squash     bool
	parallel   int
	shrinkwrap bool
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	buildCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	buildCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	buildCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	buildCmd.Flags().StringVar(&language, "lang", "", "Programming language template")

	// Setup flags that are used only by this command (variables defined above)
	buildCmd.Flags().BoolVar(&nocache, "no-cache", false, "Do not use Docker's build cache")
	buildCmd.Flags().BoolVar(&squash, "squash", false, `Use Docker's squash flag for smaller images
						 [experimental] `)
	buildCmd.Flags().IntVar(&parallel, "parallel", 1, "Build in parallel to depth specified.")

	buildCmd.Flags().BoolVar(&shrinkwrap, "shrinkwrap", false, "Just write files to ./build/ folder for shrink-wrapping")

	// Set bash-completion.
	_ = buildCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	faasCmd.AddCommand(buildCmd)
}

// buildCmd allows the user to build an OpenFaaS function container
var buildCmd = &cobra.Command{
	Use: `build -f YAML_FILE [--no-cache] [--squash]
  faas-cli build --image IMAGE_NAME
                 --handler HANDLER_DIR
                 --name FUNCTION_NAME
                 [--lang <ruby|python|python3|node|csharp|Dockerfile>]
                 [--no-cache] [--squash]
                 [--regex "REGEX"]
				 [--filter "WILDCARD"]
				 [--parallel PARALLEL_DEPTH]`,
	Short: "Builds OpenFaaS function containers",
	Long: `Builds OpenFaaS function containers either via the supplied YAML config using
the "--yaml" flag (which may contain multiple function definitions), or directly
via flags.`,
	Example: `  faas-cli build -f https://domain/path/myfunctions.yml
  faas-cli build -f ./samples.yml --no-cache
  faas-cli build -f ./samples.yml --filter "*gif*"
  faas-cli build -f ./samples.yml --regex "fn[0-9]_.*"
  faas-cli build --image=my_image --lang=python --handler=/path/to/fn/
                 --name=my_fn --squash`,
	RunE: runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	bargs := options.BuildOptions{
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
		Nocache: nocache,
		Squash: squash,
		Parallel: parallel,
		Shrinkwrap: shrinkwrap,
	}
	return api.Build(bargs)
}
