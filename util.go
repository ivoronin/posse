package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func errx(fmts string, args ...interface{}) {
	efmts := fmt.Sprintf("%s: %s\n", filepath.Base(os.Args[0]), fmts)
	fmt.Fprintf(os.Stderr, efmts, args...)
	os.Exit(1)
}

func panicf(fmts string, args ...interface{}) {
	panic(fmt.Sprintf(fmts, args...))
}
