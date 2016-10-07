package object

import (
	"flag"
	"fmt"
	fp "path/filepath"
)

type flagVal struct {
	Rename_flag string
}

func parseArgs(args []string) (*flagVal, error) {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flag to default to original filename
	rename := flagSet.String("n", fp.Base(args[1]), "Rename object before uploading")

	// Parse optional flags if they have been provided
	if len(args) > 2 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Rename_flag: string(*rename),
	}

	return &flagVals, nil
}

func main() {
	fmt.Println("put-object")
}
