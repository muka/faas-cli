package options

//BuildOptions a set of arguments used for build
type BuildOptions struct {
	FaasOptions
	SharedOptions
	Nocache    bool
	Squash     bool
	Parallel   int
	Shrinkwrap bool
}
