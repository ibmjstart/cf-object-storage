package object

import (
	"flag"
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/ncw/swift"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	sg "github.ibm.com/ckwaldon/swiftlygo"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

// argVal holds the parsed argument values.
type argVal struct {
	Container string
	source    string
	flagVals  *flagVal
}

// flagVal holds the flag values.
type flagVal struct {
	rename_flag string
}

// ParseArgs parses the arguments provided to put-object.
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

func GetObject(dest auth.Destination, container string, object string) (string, swift.Headers, error) {
	objectRet, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Object(container, object)
	if err != nil {
		return "", headers, fmt.Errorf("Failed to get object %s: %s", object, err)
	}

	return objectRet.Name, headers, nil
}

func GetObjects(dest auth.Destination, container string) ([]string, error) {
	objects, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectNamesAll(container, nil)
	if err != nil {
		return objects, fmt.Errorf("Failed to get objects: %s", err)
	}

	return objects, nil
}
