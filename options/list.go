package options

//ListOptions contains flags to list functions
type ListOptions struct {
	FaasOptions
	SharedOptions
	VerboseList bool
}
