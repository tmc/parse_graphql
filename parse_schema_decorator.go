package parse_graphql

import (
	"fmt"

	"github.com/tmc/parse"
)

// DecorateSchema modifies the provided schema and creates "Reverse" fields for Pointers.
func DecorateSchema(schema map[string]*parse.Schema) error {

	for className, classInfo := range schema {
		for fieldName, fieldInfo := range classInfo.Fields {
			if fieldInfo.Type == "Pointer" {
				fields := schema[fieldInfo.TargetClass].Fields
				reverseFieldName := fmt.Sprintf("%s_%s", className, fieldName)
				if _, alreadyPresent := fields[reverseFieldName]; alreadyPresent {
					return fmt.Errorf("Cannot create reverse pointer for %s on %v - field already present",
						reverseFieldName, fieldInfo.TargetClass)
				}
				fields[reverseFieldName] = parse.SchemaField{
					Type:        "ReversePointer",
					TargetClass: className,
				}
			}
		}
	}
	return nil
}
