package options

//DeployOptions contains flag used to deploy a function
type DeployOptions struct {
	FaasOptions
	SharedOptions
	EnvvarOpts  []string
	Replace     bool
	Update      bool
	Constraints []string
	Secrets     []string
	LabelOpts   []string
}
