package object

import (
	"flag"
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	sg "github.ibm.com/ckwaldon/swiftlygo"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

// argVal holds the parsed argument values
type argVal struct {
	Container string
	source    string
	flagVals  *flagVal
}

// flagVal holds the flag values.
type flagVal struct {
	rename_flag string
}

// parseArgs parses the arguments provided to put-object.
func ParseArgs(args []string) (*argVal, error) {
	container := args[0]
	sourceFile := args[1]

	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flag to default to original filename
	rename := flagSet.String("n", fp.Base(sourceFile), "Rename object before uploading")

	// Parse optional flags if they have been provided
	if len(args) > 2 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		rename_flag: string(*rename),
	}

	argVals := argVal{
		Container: container,
		source:    sourceFile,
		flagVals:  &flagVals,
	}

	return &argVals, nil
}

// PutObject uploads a file to Object Storage.
func PutObject(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, argVals *argVal) (string, error) {
	writer.SetCurrentStage("Uploading object")

	// Verify that the source file exists
	file, err := os.Open(argVals.source)
	if err != nil {
		return "", fmt.Errorf("Failed to open source file: %s", err)
	}
	defer file.Close()

	// Create uploader to upload object
	uploader := sg.NewObjectUploader(dest, file, argVals.Container, argVals.flagVals.rename_flag)
	err = uploader.Upload()
	if err != nil {
		return "", fmt.Errorf("Failed to upload object: %s", err)
	}

	return argVals.flagVals.rename_flag, nil
}
