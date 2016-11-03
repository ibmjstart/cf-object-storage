package command

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

type Command struct {
	cliConnection plugin.CliConnection
	parseArgs     func([]string) ([]string, error)
	cmdLogic      func([]string) error
}

func NewCommand(parseArgs func([]string) ([]string, error), cmdLogic func([]string) error) *Command {
	return &Command{
		parseArgs: parseArgs,
		cmdLogic:  cmdLogic,
	}
}

func (c *Command) Execute(args []string) error {
	// Parse arguments
	argVals, err := c.parseArgs(args)

	// Display startup info
	err = displayUserInfo(c.cliConnection)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}
}

// displayUserInfo shows the username, org and space corresponding to the requested service.
func displayUserInfo(cliConnection plugin.CliConnection) error {
	// Find username
	username, err := cliConnection.Username()
	if err != nil {
		return fmt.Errorf("Failed to get username: %s", err)
	}

	// Find org
	org, err := cliConnection.GetCurrentOrg()
	if err != nil {
		return fmt.Errorf("Failed to get organization: %s", err)
	}

	// Find space
	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return fmt.Errorf("Failed to get space: %s", err)
	}

	fmt.Printf("Working in org %s / space %s as %s...\n", cw.Cyan(org.Name), cw.Cyan(space.Name), cw.Cyan(username))

	return nil
}
