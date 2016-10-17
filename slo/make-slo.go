package slo

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
	sg "github.ibm.com/ckwaldon/swiftlygo/slo"
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
	only_missing_flag bool
	output_file_flag  string
	chunk_size_flag   int
	num_threads_flag  int
}

// ParseArgs parses the arguments provided to make-slo.
func ParseArgs(args []string) (*argVal, error) {
	sloContainer := args[0]
	sloName := args[1]
	source := args[2]

	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flags and set defaults
	missing := flagSet.Bool("m", false, "Only upload missing chunks")
	output := flagSet.String("o", "", "Destination for log data")
	chunkSize := flagSet.Int("s", -1, "Chunk size, in bytes (defaults to create 1000 chunks)")
	threads := flagSet.Int("t", runtime.NumCPU(), "Maximum number of uploader threads (defaults to the available number of CPUs")

	// Parse optional flags if they have been provided
	if len(args) > 3 {
		err := flagSet.Parse(args[3:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		only_missing_flag: bool(*missing),
		output_file_flag:  string(*output),
		chunk_size_flag:   int(*chunkSize),
		num_threads_flag:  int(*threads),
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
func MakeSlo(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, argVals *argVal) error {
	writer.SetCurrentStage("Preparing SLO")

	// Verify source file exists
	file, err := os.Open(argVals.source)
	if err != nil {
		return fmt.Errorf("Failed to open source file: %s", err)
	}
	defer file.Close()

	// Set default chunk size to create 1000 chunks if no size proveded
	if argVals.flagVals.chunk_size_flag <= 0 {
		fileStats, err := file.Stat()
		if err != nil {
			return fmt.Errorf("Failed to obtain file stats: %s", err)
		}

		argVals.flagVals.chunk_size_flag = int(math.Ceil(float64(fileStats.Size()) / 1000.0))
	}

	writer.SetCurrentStage("Uploading SLO")

	var uploader *sg.Uploader
	if argVals.flagVals.output_file_flag == "" {
		// Create SLO uploader without output file
		uploader, err = sg.NewUploader(dest, uint(argVals.flagVals.chunk_size_flag), argVals.SloContainer, argVals.SloName, file, uint(argVals.flagVals.num_threads_flag), argVals.flagVals.only_missing_flag, ioutil.Discard)
	} else {
		// Verify output file exists and create it if it does not
		outFile, err := os.OpenFile(argVals.flagVals.output_file_flag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		defer file.Close()
		if err != nil {
			return fmt.Errorf("Failed to open output file: %s", err)
		}

		// Create SLO uploader with output file
		uploader, err = sg.NewUploader(dest, uint(argVals.flagVals.chunk_size_flag), argVals.SloContainer, argVals.SloName, file, uint(argVals.flagVals.num_threads_flag), argVals.flagVals.only_missing_flag, outFile)
	}

	if err != nil {
		return fmt.Errorf("Failed to create SLO uploader: %s", err)
	}

	// Provide the console writer with upload status
	writer.SetStatus(uploader.Status)

	// Upload SLO
	err = uploader.Upload()
	if err != nil {
		return fmt.Errorf("Failed to upload SLO: %s", err)
	}

	return nil
}
