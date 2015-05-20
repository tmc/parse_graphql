package parse_graphql

import (
	"encoding/json"
	"fmt"

	"github.com/tmc/graphql"
	"github.com/tmc/graphql/executor/resolver"
	"github.com/tmc/graphql/schema"
	"github.com/tmc/parse"
)

type ParseClass struct {
	client *parse.Client
	class  *parse.Schema
	schema map[string]*parse.Schema
	Data   map[string]interface{}
}

func NewParseClass(client *parse.Client, className string, schema map[string]*parse.Schema) (*ParseClass, error) {
	class, ok := schema[className]
	if !ok {
		return nil, fmt.Errorf("class '%s' not found in schema.", className)
	}
	return &ParseClass{
		client: client,
		class:  class,
		schema: schema,
	}, nil
}

func (p *ParseClass) GraphQLTypeInfo() schema.GraphQLTypeInfo {
	className := p.class.ClassName
	ti := schema.GraphQLTypeInfo{
		Name:        p.class.ClassName,
		Description: fmt.Sprintf("Parse Class %v", className),
		Fields: schema.GraphQLFieldSpecMap{
			className: &schema.GraphQLFieldSpec{
				Name:        p.class.ClassName,
				Description: fmt.Sprintf("Root call to fetch %s", className),
				Func:        p.get,
				IsRootCall:  true,
			},
		},
	}

	// generate basic value accessors
	for fieldName, fieldSchema := range p.class.Fields {
		fn := fieldName
		ti.Fields[fieldName] = &schema.GraphQLFieldSpec{
			Name:        fn,
			Description: fmt.Sprintf("Accessor for %s field (%v)", fn, fieldSchema.Type),
			Func: func(r resolver.Resolver, f *graphql.Field) (interface{}, error) {
				partial, err := p.resolve(fn)
				if err != nil {
					return nil, err
				}
				return r.Resolve(partial, f)
			},
		}
	}
	return ti
}

func (p *ParseClass) resolve(fieldName string) (interface{}, error) {
	fieldInfo := p.class.Fields[fieldName]
	if fieldInfo.Type == "Pointer" {
		pc, err := NewParseClass(p.client, fieldInfo.TargetClass, p.schema)
		if err != nil {
			return nil, err
		}
		objectID := p.Data[fieldName].(map[string]interface{})["objectId"].(string)
		err = pc.client.GetClass(fieldInfo.TargetClass, objectID, &pc.Data)
		return pc, err
	} else {
		return p.Data[fieldName], nil
	}
}

func (p *ParseClass) get(r resolver.Resolver, f *graphql.Field) (interface{}, error) {
	var results []map[string]interface{}

	whereClause := make(map[string]interface{})
	for _, a := range f.Arguments {
		whereClause[a.Name] = a.Value
	}
	whereJSON, err := json.Marshal(whereClause)
	if err != nil {
		return nil, err
	}
	query := &parse.QueryOptions{
		Where: string(whereJSON),
	}
	if err := p.client.QueryClass(p.class.ClassName, query, &results); err != nil {
		return nil, err
	}
	typedResults := make([]*ParseClass, 0, len(results))

	for _, r := range results {
		pc, err := NewParseClass(p.client, p.class.ClassName, p.schema)
		if err != nil {
			return nil, err
		}
		pc.Data = r
		typedResults = append(typedResults, pc)
	}

	return typedResults, err
}
