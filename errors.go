package main

// ErrorBadInput is thrown whenever input data is bad in some way
// (incomplete, formatted incorrectly, corrupt, etc)
type ErrorBadInput string

func (e ErrorBadInput) Error() string {
	return string(e)
}
