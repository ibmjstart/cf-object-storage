package dlo

import (
	"flag"
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	sg "github.ibm.com/ckwaldon/swiftlygo"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

type flagVal struct {
	Container_flag string
	Prefix_flag    string
}

func parseArgs(args []string) (*flagVal, error) {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	container := flagSet.String("c", args[0], "Destination container for DLO segments (defaults to manifest container)")
	prefix := flagSet.String("p", args[1], "Prefix to be used for DLO segments (defaults to DLO name)")

	if len(args) > 2 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Container_flag: string(*container),
		Prefix_flag:    string(*prefix),
	}

	return &flagVals, nil
}

func MakeDlo(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, args []string) error {
	writer.SetCurrentStage("Preparing DLO manifest")
	flags, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("Failed to parse arguments: %s", err)
	}
	uploader := sg.NewDloManifestUploader(dest, args[0], args[1], flags.Container_flag, flags.Prefix_flag)
	writer.SetCurrentStage("Uploading DLO manifest")
	err = uploader.Upload()
	if err != nil {
		return fmt.Errorf("Failed to upload DLO manifest: %s", err)
	}
	return nil
}
