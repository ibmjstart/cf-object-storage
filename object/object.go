package object

import (
	"fmt"

	"github.com/ncw/swift"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

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

func PutObject(dest auth.Destination, container, objectName, path string) {

}
