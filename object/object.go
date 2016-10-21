package object

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ncw/swift"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

const maxObjectSize uint = 1000 * 1000 * 1000 * 5

func GetObject(dest auth.Destination, container string, object string) (string, swift.Headers, error) {
	objectRet, headers, err := dest.(*auth.SwiftDestination).SwiftConnection.Object(container, object)
	if err != nil {
		return "", headers, fmt.Errorf("Failed to get object %s: %s", object, err)
	}

	return objectRet.Name, headers, nil
}

func GetObjects(dest auth.Destination, container string) ([]string, error) {
	objects, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectNamesAll(container, nil)
	if err != nil {
		return objects, fmt.Errorf("Failed to get objects: %s", err)
	}

	return objects, nil
}

func PutObject(dest auth.Destination, container, objectName, path string) error {
	data, err := getFileContents(path)
	if err != nil {
		return fmt.Errorf("Failed to get file contents at path %s: %s", path, err)
	}

	hash := hashSource(data)

	objectCreator, err := dest.(*auth.SwiftDestination).SwiftConnection.ObjectCreate(container, objectName, true, hash, "", nil)
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
