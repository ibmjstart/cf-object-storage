package object

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	cw "github.com/ibmjstart/cf-object-storage/console_writer"
	"github.com/ibmjstart/swiftlygo/auth"
	"github.com/ncw/swift"
)

const maxObjectSize uint = 1000 * 1000 * 1000 * 5

func GetObjectInfo(dest auth.Destination, args []string) (string, error) {
	container := args[3]
	object := args[4]

	objectInfo, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Object(container, object)
	if err != nil {
		return "", fmt.Errorf("Failed to get object %s: %s", object, err)
	}

	retval := fmt.Sprintf("\r%s%s\n\nName: %s\nContent type: %s\nSize: %d bytes\nLast modified: %s\n"+
		"Hash: %s\nIs pseudo dir: %t\nSubdirectory: \n%sHeaders:", cw.ClearLine, cw.Green("OK"),
		objectInfo.Name, objectInfo.ContentType, objectInfo.Bytes, objectInfo.ServerLastModified,
		objectInfo.Hash, objectInfo.PseudoDirectory, objectInfo.SubDir)
	for k, h := range headers {
		retval += fmt.Sprintf("\n\tName: %s Value: %s", k, h)
	}
	retval += fmt.Sprintf("\n")

	return retval, nil
}

func ShowObjects(dest auth.Destination, args []string) (string, error) {
	container := args[3]

	objects, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectNamesAll(container, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get objects: %s", err)
	}

	return fmt.Sprintf("\r%s%s\n\nObjects in container %s: %v\n", cw.ClearLine, cw.Green("OK"), container, objects), nil
}

func PutObject(dest auth.Destination, container, objectName, path string, headers swift.Headers) error {
	data, err := getFileContents(path)
	if err != nil {
		return fmt.Errorf("Failed to get file contents at path %s: %s", path, err)
	}

	hash := hashSource(data)

	objectCreator, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectCreate(container, objectName, true, hash, "", headers)
	if err != nil {
		return fmt.Errorf("Failed to create object: %s", err)
	}

	_, err = objectCreator.Write(data)
	if err != nil {
		return fmt.Errorf("Failed to write object: %s", err)
	}

	err = objectCreator.Close()
	if err != nil {
		return fmt.Errorf("Failed to close object writer: %s", err)
	}

	return nil
}

func CopyObject(dest auth.Destination, container, objectName, newContainer, newName string) error {
	_, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectCopy(container, objectName, newContainer, newName, nil)
	if err != nil {
		return fmt.Errorf("Failed to rename object: %s", err)
	}

	return nil
}

func GetObject(dest auth.Destination, container, objectName, destinationPath string) error {
	object, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("Failed to open/create object file: %s", err)
	}
	defer object.Close()

	_, err = dest.(*auth.SwiftDestination).SwiftConnection.ObjectGet(container, objectName, object, true, nil)
	if err != nil {
		return fmt.Errorf("Failed to get object %s: %s", objectName, err)
	}

	return nil
}

func DeleteObject(dest auth.Destination, container, objectName string) error {
	err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectDelete(container, objectName)
	if err != nil {
		return fmt.Errorf("Failed to delete object %s: %s", objectName, err)
	}

	return nil
}

func DeleteLargeObject(dest auth.Destination, container, objectName string) error {
	// Using the Open Stack Object Storage API directly as large object support is not
	// included in the ncw/swift library yet. There is an open pull request to merge the
	// large-object branch as of 11/22/16 at https://github.com/ncw/swift/pull/74.
	var client http.Client

	authUrl := dest.(*auth.SwiftDestination).SwiftConnection.StorageUrl
	authToken := dest.(*auth.SwiftDestination).SwiftConnection.AuthToken

	deleteUrl := authUrl + "/" + container + "/" + objectName + "?multipart-manifest=delete"

	request, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return fmt.Errorf("Failed to create request: %s")
	}
	request.Header.Set("X-Auth-Token", authToken)

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Failed to make request: %s")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Failed to delete object with status %s", response.Status)
	}

	return nil
}

func getFileContents(sourcePath string) ([]byte, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open source file: %s", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Failed to get source file info: %s")
	}

	if uint(info.Size()) > maxObjectSize {
		return nil, fmt.Errorf("%s is too large to upload as a single object (max 5GB)", info.Name())
	}

	data := make([]byte, info.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to read source file: %s", err)
	}

	return data, nil
}

func hashSource(sourceData []byte) string {
	hashBytes := md5.Sum(sourceData)
	hash := hex.EncodeToString(hashBytes[:])

	return hash
}
