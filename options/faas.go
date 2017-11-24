package options

import "github.com/openfaas/faas-cli/stack"

//FaasOptions contains flags for all commands
type FaasOptions struct {
	YamlFile string
	Regex    string
	Filter   string
	WorkDir  string
	Services *stack.Services
}

//SharedOptions contains flags for subset of commands
type SharedOptions struct {
	Fprocess     string
	FunctionName string
	Network      string
	Gateway      string
	Handler      string
	Image        string
	Language     string
}
