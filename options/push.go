package options

// PushOptions store flags for the push  command
type PushOptions struct {
	FaasOptions
	SharedOptions
	Parallel int
}
