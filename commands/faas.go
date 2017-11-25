// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openfaas/faas-cli/api"
	"github.com/openfaas/faas-cli/api/template"
	"github.com/openfaas/faas-cli/options"
	"github.com/spf13/cobra"
)

// Flags that are to be added to all commands.
var (
	yamlFile string
	regex    string
	filter   string
	workDir  string
)

// Flags that are to be added to subset of commands.
var (
	fprocess     string
	functionName string
	network      string
	gateway      string
	handler      string
	image        string
	language     string
)

var stat = func(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// TODO: remove this workaround once these vars are no longer global
func resetForTest() {
	yamlFile = ""
	regex = ""
	filter = ""
}

func getFaasOptions() options.FaasOptions {
	return options.FaasOptions{
		YamlFile: yamlFile,
		Regex:    regex,
		Filter:   filter,
		WorkDir:  workDir,
	}
}

func getSharedOptions() options.SharedOptions {
	return options.SharedOptions{
		Network:      network,
		Image:        image,
		Handler:      handler,
		FunctionName: functionName,
		Language:     language,
		Gateway:      gateway,
	}
}

func init() {
	faasCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	faasCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	faasCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")
	faasCmd.PersistentFlags().StringVarP(&workDir, "workdir", "", "./", "Base directory where to store templates and build output")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = faasCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)
}

// Execute TODO
func Execute(customArgs []string) {
	checkAndSetDefaultYaml()

	faasCmd.SilenceUsage = true
	faasCmd.SilenceErrors = true
	faasCmd.SetArgs(customArgs[1:])
	if err := faasCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

func checkAndSetDefaultYaml() {
	// Check if there is a default yaml file and set it
	if _, err := stat(api.DefaultYAML); err == nil {
		yamlFile = api.DefaultYAML
	}
}

// faasCmd is the FaaS CLI root command and mimics the legacy client behaviour
// Every other command attached to FaasCmd is a child command to it.
var faasCmd = &cobra.Command{
	Use:   "faas-cli",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		err := template.SetWorkDirectory(workDir)
		if err != nil {
			log.Fatalf("Failed to set working directory: %s", err.Error())
		}

	},
	Run: runFaas,
}

// runFaas TODO
func runFaas(cmd *cobra.Command, args []string) {
	fmt.Printf(figletStr)
	cmd.Help()
}
