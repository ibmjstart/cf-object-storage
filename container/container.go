package container

import (
	"fmt"
	"strings"

	cw "github.com/ibmjstart/cf-object-storage/console_writer"
	"github.com/ibmjstart/swiftlygo/auth"
	"github.com/ncw/swift"
)

var shortHeaders = map[string]string{
	"r":  "X-Container-Read:.r:*",
	"-r": "X-Remove-Container-Read:1",
}

func ShowContainers(dest auth.Destination, args []string) (string, error) {
	serviceName := args[2]

	containers, err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerNamesAll(nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get containers: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nContainers in OS %s: %v\n", cw.ClearLine, cw.Green("OK"), serviceName, containers), nil
}

func GetContainerInfo(dest auth.Destination, container string) (swift.Container, swift.Headers, error) {
	containerRet, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Container(container)
	if err != nil {
		return containerRet, headers, fmt.Errorf("Failed to get container info for container %s: %s", container, err)
	}

	return containerRet, headers, nil
}

func MakeContainer(dest auth.Destination, container string, headers ...string) error {
	headerMap := make(map[string]string)

	for _, h := range headers {
		hFromMap, found := shortHeaders[h]
		if found {
			h = hFromMap
		}

		headerPair := strings.SplitN(h, ":", 2)
		if len(headerPair) != 2 {
			return fmt.Errorf("Unable to parse headers (must use format header-name:header-value)")
		}

		headerMap[headerPair[0]] = headerPair[1]
	}

	swiftHeader := swift.Headers(headerMap)

	err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerCreate(container, swiftHeader)
	if err != nil {
		return fmt.Errorf("Failed to create container: %s", err)
	}

	return nil
}

func DeleteContainer(dest auth.Destination, container string) error {
	err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerDelete(container)
	if err != nil {
		return fmt.Errorf("Failed to delete container: %s", err)
	}

	return nil
}
