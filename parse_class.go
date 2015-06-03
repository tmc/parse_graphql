package parse_graphql

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/graphql"
	"github.com/tmc/graphql/executor/resolver"
	"github.com/tmc/graphql/executor/tracer"
	"github.com/tmc/graphql/schema"
	"github.com/tmc/parse"
	"golang.org/x/net/context"
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
				Description: fmt.Sprintf("Root field to fetch %s", className),
				Func:        p.get,
				IsRoot:      true,
			},
		},
	}

	// generate basic value accessors
	for fieldName, fieldSchema := range p.class.Fields {
		fn := fieldName
		ti.Fields[fieldName] = &schema.GraphQLFieldSpec{
			Name:        fn,
			Description: fmt.Sprintf("Accessor for %s field (%v)", fn, fieldSchema.Type),
			Func: func(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
				partial, err := p.resolve(ctx, r, fn)
				if err != nil {
					return nil, err
				}
				return r.Resolve(ctx, partial, f)
			},
		}
	}
	return ti
}

func (p *ParseClass) resolve(ctx context.Context, r resolver.Resolver, fieldName string) (interface{}, error) {
	log.Println("resolving ===============================================")
	fieldInfo := p.class.Fields[fieldName]
	if fieldInfo.Type == "Pointer" {
		pc, err := NewParseClass(p.client, fieldInfo.TargetClass, p.schema)
		if err != nil {
			return nil, err
		}
		objectID := p.Data[fieldName].(map[string]interface{})["objectId"].(string)
		if t, ok := tracer.FromContext(ctx); ok {
			t.IncQueries(1)
		}
		err = pc.client.GetClass(fieldInfo.TargetClass, objectID, &pc.Data)
		return pc, err
	} else if fieldInfo.Type == "ReversePointer" {
		pc, err := NewParseClass(p.client, fieldInfo.TargetClass, p.schema)
		if err != nil {
			return nil, err
		}
		fieldName := fieldName[len(fieldInfo.TargetClass)+1:]
		// TODO(tmc): optimize
		query := []byte(fmt.Sprintf(`{"__type":"Pointer","className":"%s","objectId":"%s"}`,
			p.class.ClassName,
			p.Data["objectId"]))
		queryEncoded := json.RawMessage(query)
		return pc.get(ctx, r, &graphql.Field{
			Arguments: []graphql.Argument{
				{
					Name:  fieldName,
					Value: &queryEncoded,
				},
			},
		})
	} else {
		return p.Data[fieldName], nil
	}
}

func (p *ParseClass) get(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
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
	if t, ok := tracer.FromContext(ctx); ok {
		t.IncQueries(1)
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
