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

// flagVal holds the flag values
type flagVal struct {
	Only_missing_flag bool
	Output_file_flag  string
	Chunk_size_flag   int
	Num_threads_flag  int
}

// parseArgs parses the arguments provided to make-slo
func parseArgs(args []string) (*flagVal, error) {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flags and set defaults
	missing := flagSet.Bool("m", false, "Only upload missing chunks")
	output := flagSet.String("o", "", "Destination for log data")
	chunkSize := flagSet.Int("s", -1, "Chunk size, in bytes (defaults to create 1000 chunks)")
	threads := flagSet.Int("t", runtime.NumCPU(), "Maximum number of uploader threads (defaults to the available number of CPUs")

	// Parse optional flags if they have been provided
	if len(args) > 4 {
		err := flagSet.Parse(args[3:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Only_missing_flag: bool(*missing),
		Output_file_flag:  string(*output),
		Chunk_size_flag:   int(*chunkSize),
		Num_threads_flag:  int(*threads),
	}

	return &flagVals, nil
}

// MakeSlo uploads the given file as an SLO to Object Storage
func MakeSlo(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, args []string) (string, error) {
	writer.SetCurrentStage("Preparing SLO")
	flags, err := parseArgs(args)
	if err != nil {
		return "", fmt.Errorf("Failed to parse arguments: %s", err)
	}

	// Verify source file exists
	file, err := os.Open(args[2])
	defer file.Close()
	if err != nil {
		return "", fmt.Errorf("Failed to open source file: %s", err)
	}

	// Set default chunk size to create 1000 chunks if no size proveded
	if flags.Chunk_size_flag <= 0 {
		fileStats, err := file.Stat()
		if err != nil {
			return "", fmt.Errorf("Failed to obtain file stats: %s", err)
		}

		flags.Chunk_size_flag = int(math.Ceil(float64(fileStats.Size()) / 1000.0))
	}

	writer.SetCurrentStage("Uploading SLO")

	var uploader *sg.Uploader
	if flags.Output_file_flag == "" {
		// Create SLO uploader without output file
		uploader, err = sg.NewUploader(dest, uint(flags.Chunk_size_flag), args[0], args[1], file, uint(flags.Num_threads_flag), flags.Only_missing_flag, ioutil.Discard)
	} else {
		// Verify output file exists and create it if it does not
		outFile, err := os.OpenFile(flags.Output_file_flag, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		defer file.Close()
		if err != nil {
			return "", fmt.Errorf("Failed to open output file: %s", err)
		}

		// Create SLO uploader with output file
		uploader, err = sg.NewUploader(dest, uint(flags.Chunk_size_flag), args[0], args[1], file, uint(flags.Num_threads_flag), flags.Only_missing_flag, outFile)
	}

	if err != nil {
		return "", fmt.Errorf("Failed to create SLO uploader: %s", err)
	}

	// Upload SLO
	err = uploader.Upload()
	if err != nil {
		return "", fmt.Errorf("Failed to upload SLO: %s", err)
	}

	return "", err
}
