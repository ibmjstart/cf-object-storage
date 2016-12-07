package slo

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	w "github.com/ibmjstart/cf-object-storage/writer"
	"github.com/ibmjstart/swiftlygo/auth"
	sg "github.com/ibmjstart/swiftlygo/slo"
)

// argVal holds the parsed argument values.
type argVal struct {
	SloContainer string
	SloName      string
	source       string
	flagVals     *flagVal
}

// flagVal holds the flag values.
type flagVal struct {
	onlyMissingFlag bool
	outputFileFlag  string
	chunkSizeFlag   int
	numThreadsFlag  int
}

// parseArgs parses the arguments provided to make-slo.
func parseArgs(args []string) (*argVal, error) {
	sloContainer := args[0]
	sloName := args[1]
	source := args[2]

	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flags and set defaults
	missing := flagSet.Bool("m", false, "Only upload missing chunks")
	output := flagSet.String("o", "", "Destination for log data")
	chunkSize := flagSet.Int("s", 1*1000*1000*1000, "Chunk size, in bytes (defaults to create 1GB chunks)")
	threads := flagSet.Int("t", runtime.NumCPU(), "Maximum number of uploader threads (defaults to the available number of CPUs")

	// Parse optional flags if they have been provided
	if len(args) > 3 {
		err := flagSet.Parse(args[3:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		onlyMissingFlag: bool(*missing),
		outputFileFlag:  string(*output),
		chunkSizeFlag:   int(*chunkSize),
		numThreadsFlag:  int(*threads),
	}

	argVals := argVal{
		SloContainer: sloContainer,
		SloName:      sloName,
		source:       source,
		flagVals:     &flagVals,
	}

	return &argVals, nil
}

// MakeSlo uploads the given file as an SLO to Object Storage.
func MakeSlo(dest auth.Destination, writer *w.ConsoleWriter, args []string) (string, error) {
	writer.SetCurrentStage("Preparing SLO")

	argVals, err := parseArgs(args[3:])

	// Verify source file exists
	file, err := os.Open(argVals.source)
	if err != nil {
		return "", fmt.Errorf("Failed to open source file: %s", err)
	}
	defer file.Close()

	fileStats, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("Failed to obtain file stats: %s", err)
	}

	if int(fileStats.Size()) < argVals.flagVals.chunkSizeFlag {
		argVals.flagVals.chunkSizeFlag = int(fileStats.Size())
	}

	writer.SetCurrentStage("Uploading SLO")

	var output io.Writer
	if argVals.flagVals.outputFileFlag == "" {
		output = ioutil.Discard
	} else {
		// Verify output file exists and create it if it does not
		output, err = os.OpenFile(argVals.flagVals.outputFileFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		defer file.Close()
		if err != nil {
			return "", fmt.Errorf("Failed to open output file: %s", err)
		}
	}

	// Create SLO uploader without output file
	uploader, err := sg.NewUploader(dest, uint(argVals.flagVals.chunkSizeFlag),
		argVals.SloContainer, argVals.SloName, file, uint(argVals.flagVals.numThreadsFlag),
		argVals.flagVals.onlyMissingFlag, output)
	if err != nil {
		return "", fmt.Errorf("Failed to create SLO uploader: %s", err)
	}

	// Provide the console writer with upload status
	writer.SetStatus(uploader.Status)

	// Upload SLO
	err = uploader.Upload()
	if err != nil {
		return "", fmt.Errorf("Failed to upload SLO: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n%s\nSuccessfully created SLO %s in container %s\n", w.ClearLine, w.Green("OK"), w.ClearLine, w.Cyan(argVals.SloName), w.Cyan(argVals.SloContainer)), nil
}
