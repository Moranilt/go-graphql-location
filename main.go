package main

import (
	"log"
	"net/http"

	"github.com/Moranilt/go-graphql-location/objects"
	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/graphql-go-handler"
)

var Query = graphql.NewObject(graphql.ObjectConfig{
	Name: "GeoLocationRemember",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewObject(objects.User),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return objects.User, nil
			},
		},
		"location": &graphql.Field{
			Type: graphql.NewObject(objects.Location),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return objects.Location, nil
			},
		},
	},
})

func main() {
	var schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: Query,
	})
	if err != nil {
		log.Fatalf("Failed to create new schema. Error: %v", err)
	}
	handler := gqlhandler.New(&gqlhandler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
