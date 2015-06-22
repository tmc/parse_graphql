package parse_graphql

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/tmc/graphql"
	"github.com/tmc/graphql/executor/resolver"
	"github.com/tmc/graphql/executor/tracer"
	"github.com/tmc/graphql/schema"
	"github.com/tmc/parse"
	"golang.org/x/net/context"
)

var DefaultLimit = 5

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
				partial, err := p.resolve(ctx, r, f)
				if err != nil {
					return nil, err
				}
				return r.Resolve(ctx, partial, f)
			},
		}
	}
	return ti
}

func (p *ParseClass) resolve(ctx context.Context, r resolver.Resolver, field *graphql.Field) (interface{}, error) {
	log.Println("resolving ===============================================")
	fieldInfo := p.class.Fields[field.Name]
	if fieldInfo.Type == "Pointer" {
		return p.resolvePointer(ctx, r, field)
	} else if fieldInfo.Type == "ReversePointer" {
		return p.resolveReversePointer(ctx, r, field)
	} else if fieldInfo.Type == "HookFunction" {
		return mkHookFieldFunc(p.client, p.schema, p.class.ClassName+"_"+field.Name, p.Data)(ctx, r, field)
	} else {
		return p.Data[field.Name], nil
	}
}

func (p *ParseClass) resolvePointer(ctx context.Context, r resolver.Resolver, field *graphql.Field) (interface{}, error) {
	fieldName := field.Name
	fieldInfo := p.class.Fields[fieldName]
	pc, err := NewParseClass(p.client, fieldInfo.TargetClass, p.schema)
	if err != nil {
		return nil, err
	}
	// only resolve if we can obtain the objectId
	data, ok := p.Data[fieldName]
	if !ok {
		return nil, nil
	}
	asMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	oid := asMap["objectId"]
	if oid == nil {
		return nil, nil
	}
	objectID, ok := oid.(string)
	if !ok {
		return nil, nil
	}
	if t, ok := tracer.FromContext(ctx); ok {
		t.IncQueries(1)
	}
	err = pc.client.GetClass(fieldInfo.TargetClass, objectID, &pc.Data)
	return pc, err
}

func (p *ParseClass) resolveReversePointer(ctx context.Context, r resolver.Resolver, field *graphql.Field) (interface{}, error) {
	fieldInfo := p.class.Fields[field.Name]
	pc, err := NewParseClass(p.client, fieldInfo.TargetClass, p.schema)
	if err != nil {
		return nil, err
	}
	fieldName := field.Name[len(fieldInfo.TargetClass)+1:]
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
}

var specialFields = []string{"order", "limit", "skip", "keys", "include", "where"}
var specialFieldsSet map[string]bool

func (p *ParseClass) get(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
	var results []map[string]interface{}

	// TODO(tmc): handle overlap between special fields and user defined fields on a class elegantly
	whereClause := make(map[string]interface{})
	for _, a := range f.Arguments {
		// only populate where clause if the field isn't in out special field list
		if !specialFieldsSet[a.Name] {
			whereClause[a.Name] = a.Value
		}
	}
	if explicitWhere, ok := f.Arguments.Get("where"); ok {
		asMap, ok := explicitWhere.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("explicit where fields must be maps, got '%T'", explicitWhere)
		}
		whereClause = asMap
	}
	whereJSON, err := json.Marshal(whereClause)
	if err != nil {
		return nil, err
	}
	// limit
	limit := DefaultLimit
	if l, ok := f.Arguments.Get("limit"); ok {
		if lim, ok := l.(int); ok {
			limit = lim
		} else {
			return nil, fmt.Errorf("'limit' argument should be an integer. Got %#v", l)
		}
	}

	// order
	order := ""
	if o, ok := f.Arguments.Get("order"); ok {
		if orderStr, ok := o.(string); ok {
			order = orderStr
		} else {
			return nil, fmt.Errorf("'order' argument should be a string. Got %#v", o)
		}
	}
	spew.Dump("ORDER:", order, f.Arguments)
	query := &parse.QueryOptions{
		Where: string(whereJSON),
		Limit: limit,
		Order: order,
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

func init() {
	specialFieldsSet = make(map[string]bool)
	for _, f := range specialFields {
		specialFieldsSet[f] = true
	}
}
