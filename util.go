package main

import (
	"fmt"
	"os"
)

func errx(fmts string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmts, args...)
	os.Exit(1)
}

func panicf(fmts string, args ...interface{}) {
	panic(fmt.Sprintf(fmts, args...))
}
