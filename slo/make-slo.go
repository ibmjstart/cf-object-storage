package slo

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	sg "github.ibm.com/ckwaldon/swiftlygo"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
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
	if len(args) > 3 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Only_missing_flag: bool(*missing),
		Output_file_flag:  string(*output),
		Chunk_size_flag:   strconv.Itoa(*chunkSize),
		Num_threads_flag:  strconv.Itoa(*threads),
	}

	return &flagVals, nil
}

func main() {
	fmt.Println("slo")
}
