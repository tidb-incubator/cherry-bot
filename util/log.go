package util

import (
	"log"

	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// Error print error stack
func Error(err error) {
	if err == nil {
		return
	}
	e := errors.Cause(err)

	if st, ok := e.(stackTracer); ok {
		log.Printf("%+v\n", st.StackTrace()[:])
	} else {
		log.Println(err)
	}
}

// Event print event
func Event(a ...interface{}) {
	log.Println(a...)
}

// Println wrap standard log library function
func Println(a ...interface{}) {
	log.Println(a...)
}

// Printf wrap standard log library function
func Printf(l string, a ...interface{}) {
	log.Printf(l, a...)
}

// Fatal print error info and exit process
func Fatal(err error) {
	if err == nil {
		return
	}
	e := errors.Cause(err)

	if st, ok := e.(stackTracer); ok {
		log.Fatal(st.StackTrace()[:])
	} else {
		log.Fatal(err)
	}
}
