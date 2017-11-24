package options

//InvokeOptions contains flags to invoke a function
type InvokeOptions struct {
	FaasOptions
	SharedOptions
	ContentType  string
	Query        []string
	FunctionName string
	Input        []byte
}
