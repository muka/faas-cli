package options

//LoginOptions contains flags to login
type LoginOptions struct {
	SharedOptions
	Username string
	Password string
}
