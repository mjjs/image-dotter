package main

import (
	"fmt"
	"log"
	"os"
)

// getSourceFilename parses the arguments of the program and returns the filename if found.
// If a filename cannot be found in the arguments, usage info is printed.
func getSourceFilename() string {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("Usage: image-shaper filename")
	}
	return args[0]
}

func getDestFilename() string {
	return fmt.Sprintf("out_%s", getSourceFilename())
}

func getDestinationFile() (*os.File, error) {
	return os.OpenFile(getDestFilename(), os.O_WRONLY|os.O_CREATE, 0600)
}
