package parse_graphql

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmc/graphql"
	"github.com/tmc/graphql/executor/resolver"
	"github.com/tmc/graphql/schema"
	"github.com/tmc/parse"
	"golang.org/x/net/context"
)

type ParseSchema struct {
	client *parse.Client
	Schema map[string]*parse.Schema
	// hooks that don't appear to be associated with a class
	// based on the naming scheme '<className>_Foobar'
	hooks []*parse.HookFunction
}

// NewParseSchema modifies the provided schema and creates "Reverse" fields for Pointers and
// attaches appropriately named hook functions.
func NewParseSchema(client *parse.Client, schema map[string]*parse.Schema, hooks []*parse.HookFunction) (*ParseSchema, error) {
	result := &ParseSchema{
		client: client,
		Schema: make(map[string]*parse.Schema, len(schema)),
		hooks:  make([]*parse.HookFunction, 0),
	}
	classHooks := map[string][]string{}
	for _, hook := range hooks {
		parts := strings.Split(hook.FunctionName, "_")
		if class, ok := schema[parts[0]]; ok {
			cn := class.ClassName
			if classHooks[cn] == nil {
				classHooks[cn] = []string{}
			}
			classHooks[cn] = append(classHooks[cn], strings.Join(parts[1:], "_"))
		} else {
			// if it doesn't appear associated with a type, put it on the schema object
			result.hooks = append(result.hooks, hook)
		}
	}

	for className, classInfo := range schema {
		result.Schema[className] = &(*(classInfo))
	}
	for className, classInfo := range result.Schema {
		for fieldName, fieldInfo := range classInfo.Fields {
			if fieldInfo.Type == "Pointer" {
				fields := schema[fieldInfo.TargetClass].Fields
				reverseFieldName := fmt.Sprintf("%s_%s", className, fieldName)
				if _, alreadyPresent := fields[reverseFieldName]; alreadyPresent {
					return nil, fmt.Errorf("Cannot create reverse pointer for %s on %v - field already present",
						reverseFieldName, fieldInfo.TargetClass)
				}
				fields[reverseFieldName] = parse.SchemaField{
					Type:        "ReversePointer",
					TargetClass: className,
				}
			}
		}
		if hooks, ok := classHooks[className]; ok {
			for _, hookName := range hooks {
				classInfo.Fields[hookName] = parse.SchemaField{
					Type: "HookFunction",
				}
			}
		}
	}
	return result, nil
}

func (s *ParseSchema) GraphQLTypeInfo() schema.GraphQLTypeInfo {
	ti := schema.GraphQLTypeInfo{
		Name:        "ParseSchema",
		Description: "Parse schema object",
		Fields: map[string]*schema.GraphQLFieldSpec{
			"signUp": {"signUp", "Sign up a new user.", s.signUp, []graphql.Argument{
				{Name: "username"}, {Name: "password"}, {Name: "email"},
			}, true},
			"logIn": {"logIn", "Authenticate as a user.", s.logIn, []graphql.Argument{
				{Name: "username"}, {Name: "password"},
			}, true},
			"me": {"me", "Return the currently authenticated user.", s.me, nil, true},
		},
	}

	for _, hookFunction := range s.hooks {
		hookName := hookFunction.FunctionName
		ti.Fields[hookName] = &schema.GraphQLFieldSpec{
			Name:        hookName,
			Description: fmt.Sprintf("Cloud Code function %s", hookName),
			Func:        mkHookFieldFunc(s.client, s.Schema, hookName, nil),
			IsRoot:      true,
		}
	}

	return ti
}

func (s *ParseSchema) signUp(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
	var u parse.ParseUser
	// username
	userName, ok := f.Arguments.Get("username")
	if !ok {
		return nil, fmt.Errorf("'username' field is required.")
	}
	u.Username, ok = userName.(string)
	if !ok {
		return nil, fmt.Errorf("'username' field must be a string.")
	}

	// password
	password, ok := f.Arguments.Get("password")
	if !ok {
		return nil, fmt.Errorf("'password' field is required.")
	}
	u.Password, ok = password.(string)
	if !ok {
		return nil, fmt.Errorf("'password' field must be a string.")
	}

	// email
	email, ok := f.Arguments.Get("email")
	if !ok {
		return nil, fmt.Errorf("'email' field is required.")
	}
	u.Email, ok = email.(string)
	if !ok {
		return nil, fmt.Errorf("'email' field must be a string.")
	}

	return s.client.CreateUser(u)
}

func (s *ParseSchema) logIn(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
	// username
	usernamei, ok := f.Arguments.Get("username")
	if !ok {
		return nil, fmt.Errorf("'username' field is required.")
	}
	username, ok := usernamei.(string)
	if !ok {
		return nil, fmt.Errorf("'username' field must be a string.")
	}

	// password
	passwordi, ok := f.Arguments.Get("password")
	if !ok {
		return nil, fmt.Errorf("'password' field is required.")
	}
	password, ok := passwordi.(string)
	if !ok {
		return nil, fmt.Errorf("'password' field must be a string.")
	}

	var u parse.ParseUser
	err := s.client.LoginUser(username, password, &u)
	return u, err
}

// client returns a *parse.Client that may be authed as a user if the provided context
// contains an http request with the X-Parse-Session-Token header set.
func (s *ParseSchema) authedClient(ctx context.Context) *parse.Client {
	r := ctx.Value("http_request")
	if r != nil {
		request, ok := r.(*http.Request)
		if ok && request.Header.Get("X-Parse-Session-Token") != "" {
			return s.client.WithSessionToken(request.Header.Get("X-Parse-Session-Token"))
		}
	}
	return s.client
}

func (s *ParseSchema) me(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
	c := s.authedClient(ctx)
	var user parse.ParseUser
	err := c.CurrentUser(&user)
	if err != nil {
		return nil, err
	}
	pc, err := NewParseClass(c, "_User", s.Schema)
	if err != nil {
		return nil, err
	}
	pc.Data, err = tomap(user)
	return pc, err
}

func mkHookFieldFunc(client *parse.Client, schema map[string]*parse.Schema, hookName string, data map[string]interface{}) schema.GraphQLFieldFunc {
	return func(ctx context.Context, r resolver.Resolver, f *graphql.Field) (interface{}, error) {
		output, err := client.CallCloudFunction(hookName, data)
		if err != nil {
			return nil, err
		}
		var result struct {
			Result interface{} `json:"result"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			return string(output), err
		}
		if objs, ok := result.Result.([]interface{}); ok {
			result := []interface{}{}
			for _, obj := range objs {
				mapobj, ok := obj.(map[string]interface{})
				if !ok {
					result = append(result, obj)
					continue
				}
				if cn := mapobj["className"]; cn == nil {
					result = append(result, obj)
					continue
				}
				cn, ok := mapobj["className"].(string)
				if !ok {
					result = append(result, obj)
					continue
				}
				pc, err := NewParseClass(client, cn, schema)
				pc.Data = mapobj
				if err != nil {
					result = append(result, obj)
				} else {
					result = append(result, pc)
				}
			}
			return result, nil
		}
		return result.Result, nil

	}
}

// tomap attempts to convert a value to a map[string]interface via encoding/json
func tomap(value interface{}) (map[string]interface{}, error) {
	asjson, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var asmap map[string]interface{}
	return asmap, json.Unmarshal(asjson, &asmap)
}
