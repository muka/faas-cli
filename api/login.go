package api

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/config"
	"github.com/openfaas/faas-cli/options"
)

//Login to a OpenFaaS gateway
func Login(arg options.LoginOptions) error {

	gateway := strings.TrimRight(strings.TrimSpace(arg.Gateway), "/")
	if err := validateLogin(gateway, arg.Username, arg.Password); err != nil {
		return err
	}

	if err := config.UpdateAuthConfig(gateway, arg.Username, arg.Password); err != nil {
		return err
	}

	user, _, err := config.LookupAuthConfig(gateway)
	if err != nil {
		return err
	}
	fmt.Println("credentials saved for", user, gateway)

	return nil
}

func validateLogin(url string, user string, pass string) error {
	// TODO: provide --insecure flag for this
	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(5 * time.Second),
	}

	// TODO: implement ping in the gateway API and call that
	gatewayURL := strings.TrimRight(url, "/")
	req, _ := http.NewRequest("GET", gatewayURL+"/system/functions", nil)
	req.SetBasicAuth(user, pass)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gatewayURL)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.TLS == nil {
		fmt.Println("WARNING! Communication is not secure, please consider using HTTPS. Letsencrypt.org offers free SSL/TLS certificates.")
	}

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("unable to login, either username or password is incorrect")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return nil
}
