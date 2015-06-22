package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)
import (
	"github.com/tmc/graphql/executor"
	"github.com/tmc/graphql/handler"
	"github.com/tmc/graphql/schema"
	"github.com/tmc/parse"
	"github.com/tmc/parse_graphql"
)

type ServeOptions struct {
	ListenAddr         string `short:"l" long:"listen" description:"Listen address" default:":8080"`
	ParseApplicationID string `short:"a" long:"appID" description:"Parse Application ID" env:"PARSE_APPLICATION_ID"`
	ParseMasterKey     string `short:"m" long:"masterKey" description:"Parse Master Key" env:"PARSE_MASTER_KEY"`
	ParseRESTAPIKey    string `short:"w" long:"restApiKey" description:"Parse REST API Key" env:"PARSE_REST_API_KEY"`
}

var serveOptions ServeOptions

func init() {
	if _, err := optionsParser.AddCommand("serve", "", "", &serveOptions); err != nil {
		log.Fatal(err)
	}
}

func (c *ServeOptions) Execute(args []string) error {
	log.Println(c)
	schema := schema.New()

	client, err := parse.NewClient(c.ParseApplicationID, c.ParseRESTAPIKey)
	if err != nil {
		return err
	}
	mClient := client.WithMasterKey(c.ParseMasterKey)
	client.TraceOn(log.New(os.Stdout, "[parse] ", log.LstdFlags))
	classes, err := mClient.GetFullSchema()
	if err != nil {
		return fmt.Errorf("error fetching parse app schema: %v", err)
	}
	hooks, err := mClient.GetHookFunctions()
	if err != nil {
		return fmt.Errorf("error fetching parse app hooks: %v", err)
	}
	parseSchema, err := parse_graphql.NewParseSchema(client, classes, hooks)
	if err != nil {
		return err
	}
	for _, class := range parseSchema.Schema {
		parseClass, err := parse_graphql.NewParseClass(client, class.ClassName, parseSchema.Schema)
		if err != nil {
			return err
		}
		schema.Register(parseClass)
	}
	schema.Register(parseSchema) // for top-level fields
	executor := executor.New(schema)

	mux := http.NewServeMux()
	mux.Handle("/", handler.New(executor))
	return http.ListenAndServe(c.ListenAddr, mux)
}
