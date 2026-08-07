package utils

import (
	"fmt"
	"log"
	"runtime/debug"
)

// RecoverPanic recovers a panic in the deferring function and converts it into
// an error assigned to *errp, logging the stack trace under label.
//
// Must be used as `defer utils.RecoverPanic(label, &err)` in a function with a
// named error return — recover() only stops a panic when called directly by
// the deferred function itself, so wrapping this call in another closure
// (e.g. `defer func() { utils.RecoverPanic(...) }()`) will not catch anything.
func RecoverPanic(label string, errp *error) {
	if p := recover(); p != nil {
		log.Printf("%s panic: %v\n%s", label, p, debug.Stack())
		*errp = fmt.Errorf("%s panic: %v", label, p)
	}
}
