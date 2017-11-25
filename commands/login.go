// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/api"
	"github.com/spf13/cobra"
)

var (
	username      string
	password      string
	passwordStdin bool
)

func init() {
	loginCmd.Flags().StringVarP(&gateway, "gateway", "g", api.DefaultGateway, "Gateway URL starting with http(s)://")
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Gateway username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Gateway password")
	loginCmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Reads the gateway password from stdin")

	faasCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   `login [--username USERNAME] [--password PASSWORD] [--gateway GATEWAY_URL]`,
	Short: "Log in to OpenFaaS gateway",
	Long:  "Log in to OpenFaaS gateway.\nIf no gateway is specified, the default local one will be used.",
	Example: `  faas-cli login -u user -p password --gateway http://localhost:8080
  cat ~/faas_pass.txt | faas-cli login -u user --password-stdin --gateway https://openfaas.mydomain.com`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {

	if len(username) == 0 {
		return fmt.Errorf("must provide --username or -u")
	}

	if len(password) > 0 {
		fmt.Println("WARNING! Using --password is insecure, consider using: cat ~/faas_pass.txt | faas-cli login -u user --password-stdin")
		if passwordStdin {
			return fmt.Errorf("--password and --password-stdin are mutually exclusive")
		}

		if len(username) == 0 {
			return fmt.Errorf("must provide --username with --password")
		}
	}

	if passwordStdin {
		if len(username) == 0 {
			return fmt.Errorf("must provide --username with --password-stdin")
		}

		passwordStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		password = strings.TrimSpace(string(passwordStdin))
	}

	password = strings.TrimSpace(password)
	if len(password) == 0 {
		return fmt.Errorf("must provide a non-empty password via --password or --password-stdin")
	}

	fmt.Println("Calling the OpenFaaS server to validate the credentials...")

	return api.Login(options.LoginOptions{
		SharedOptions: getSharedOptions(),
		Username: username,
		Password: password,
	})
}
