package main

// usageError is a command-line mistake: printed with the usage, exit code 2.
type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }
