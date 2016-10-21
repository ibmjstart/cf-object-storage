package container

import (
	"fmt"
	"strings"

	"github.com/ncw/swift"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

func GetContainerInfo(dest auth.Destination, container string) (string, swift.Headers, error) {
	containerRet, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Container(container)
	if err != nil {
		return "", headers, fmt.Errorf("Failed to get container info for container %s: %s", container, err)
	}

	return containerRet.Name, headers, nil
}

func ShowContainers(dest auth.Destination) ([]string, error) {
	containers, err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerNamesAll(nil)
	if err != nil {
		return containers, fmt.Errorf("Failed to get containers: %s", err)
	}

	return containers, nil
}

func MakeContainer(dest auth.Destination, container string, headers ...string) error {
	headerMap := make(map[string]string)

	for _, h := range headers {
		headerPair := strings.Split(h, ":")
		if len(headerPair) != 2 {
			return fmt.Errorf("Unable to parse headers (must use format header-name:header-value)")
		}

		fmt.Printf("\nHeader: %s Value: %s", headerPair[0], headerPair[1])
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
