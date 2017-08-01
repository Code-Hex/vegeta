package vegeta

type causer interface {
	Cause() error
}

type exiter interface {
	ExitCode() int
}

// UnwrapErrors get important message from wrapped error message
func UnwrapErrors(err error) (int, error) {
	for e := err; e != nil; {
		switch e.(type) {
		case exiter:
			return e.(exiter).ExitCode(), e
		case causer:
			e = e.(causer).Cause()
		default:
			return 1, e // default error
		}
	}
	return 0, nil
}
