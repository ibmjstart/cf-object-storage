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

func GetContainerInfo(dest auth.Destination, args []string) (string, error) {
	container := args[3]
	containerInfo, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Container(container)
	if err != nil {
		return "", fmt.Errorf("Failed to get container info for container %s: %s", container, err)
	}

	retval := fmt.Sprintf("\r%s%s\n\nName: %s\nnumber of objects: %d\nSize: %d bytes\nHeaders:", cw.ClearLine, cw.Green("OK"), containerInfo.Name, containerInfo.Count, containerInfo.Bytes)
	for k, h := range headers {
		retval += fmt.Sprintf("\n\tName: %s Value: %s", k, h)
	}
	retval += fmt.Sprintf("\n")

	return retval, nil
}

func MakeContainer(dest auth.Destination, args []string) (string, error) {
	headerMap := make(map[string]string)
	serviceName := args[2]
	container := args[3]
	headers := args[4:]

	for _, h := range headers {
		hFromMap, found := shortHeaders[h]
		if found {
			h = hFromMap
		}

		headerPair := strings.SplitN(h, ":", 2)
		if len(headerPair) != 2 {
			return "", fmt.Errorf("Unable to parse headers (must use format header-name:header-value)")
		}

		headerMap[headerPair[0]] = headerPair[1]
	}

	swiftHeader := swift.Headers(headerMap)

	err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerCreate(container, swiftHeader)
	if err != nil {
		return "", fmt.Errorf("Failed to create container: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nCreated container %s in OS %s\n", cw.ClearLine, cw.Green("OK"), container, serviceName), nil
}

func DeleteContainer(dest auth.Destination, args []string) (string, error) {
	serviceName := args[2]
	container := args[3]

	if len(args) == 5 && args[4] == "-f" {
		objects, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectNamesAll(container, nil)
		if err != nil {
			return "", fmt.Errorf("Failed to get objects to delete: %s", err)
		}

		for _, rmObject := range objects {
			err = dest.(*auth.SwiftDestination).SwiftConnection.ObjectDelete(container, rmObject)
			if err != nil {
				return "", fmt.Errorf("Failed to delete object %s: %s", rmObject, err)
			}
		}
	}

	err := dest.(*auth.SwiftDestination).SwiftConnection.ContainerDelete(container)
	if err != nil {
		return "", fmt.Errorf("Failed to delete container: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nDeleted container %s from OS %s\n", cw.ClearLine, cw.Green("OK"), container, serviceName), nil
}

func UpdateContainer(dest auth.Destination, args []string) (string, error) {
	serviceName := args[2]
	container := args[3]

	_, err := GetContainerInfo(dest, args)
	if err != nil {
		return "", fmt.Errorf("Failed to get container %s: %s", container, err)
	}

	_, err = MakeContainer(dest, args)
	if err != nil {
		return "", fmt.Errorf("Failed to make container: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nUpdated container %s in OS %s\n", cw.ClearLine, cw.Green("OK"), container, serviceName), nil
}

func RenameContainer(dest auth.Destination, args []string) (string, error) {
	container := args[3]
	newContainer := args[4]

	_, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Container(container)
	if err != nil {
		return "", fmt.Errorf("Failed to get container %s: %s", container, err)
	}

	headersArg := make([]string, 0)
	for header, val := range headers {
		headersArg = append(headersArg, header+":"+val)
	}

	makeArg := append(args[:3], append([]string{newContainer}, headersArg...)...)
	_, err = MakeContainer(dest, makeArg)
	if err != nil {
		return "", fmt.Errorf("Failed to make container: %s", err)
	}

	objects, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectNamesAll(container, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get objects to move: %s", err)
	}

	for _, mvObject := range objects {
		_, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectCopy(container, mvObject, newContainer, mvObject, nil)
		if err != nil {
			return "", fmt.Errorf("Failed to move object %s: %s", mvObject, err)
		}

		err = dest.(*auth.SwiftDestination).SwiftConnection.ObjectDelete(container, mvObject)
		if err != nil {
			return "", fmt.Errorf("Failed to delete object %s: %s", mvObject, err)
		}
	}

	_, err = DeleteContainer(dest, args[:4])
	if err != nil {
		return "", fmt.Errorf("Failed to delete container: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nRenamed container %s to %s\n", cw.ClearLine, cw.Green("OK"), container, newContainer), nil
}
