package parse

import (
	"fmt"
	"reflect"
	"regexp"
)

// ClassNamer is an interface that allows a type to provide its associated
// Parse.com object class name.
type ClassNamer interface {
	ParseClassName() string
}

func objectTypeNameFromSlice(objects interface{}) (string, error) {
	rv := reflect.ValueOf(objects)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return "", fmt.Errorf("expected slice but got %s", rv.Kind())
	}
	return objectTypeName(reflect.Zero(rv.Type().Elem()).Interface())
}

func objectTypeName(object interface{}) (string, error) {
	if namer, ok := object.(ClassNamer); ok {
		return namer.ParseClassName(), nil
	}
	rv := reflect.ValueOf(object)
	var typeName string
	switch rv.Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		typeName = rv.Type().Elem().Name()
	case reflect.Struct:
		typeName = rv.Type().Name()
	default:
		return "", fmt.Errorf("Expected a pointer or an interface type. Got %s", rv.Kind())
	}
	return typeName, nil
}

var objectURIRe = regexp.MustCompile("1/classes/(?P<className>[^/]+)(/(?P<objectID>.+))?")

func objectURIToClassAndID(uri string) (className string, objectID string) {
	matches := objectURIRe.FindStringSubmatch(uri)
	if len(matches) > 1 {
		className = matches[1]
	}
	if len(matches) > 3 {
		objectID = matches[3]
	}
	return
}
