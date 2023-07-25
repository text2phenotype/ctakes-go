package utils

import "fmt"

func RecoverWithError(err *error) {
	if rv := recover(); rv != nil {
		*err = fmt.Errorf("got panic: %v", rv)
	}
}
