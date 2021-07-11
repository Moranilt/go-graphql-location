package objects

import "github.com/graphql-go/graphql"

var Location = graphql.ObjectConfig{
	Name:        "Location",
	Description: "Get Current Location info",
	Fields: graphql.Fields{
		"coordinates": &graphql.Field{
			Type: graphql.NewObject(Coordinates),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return Coordinates, nil
			},
		},
		"place": &graphql.Field{
			Type: graphql.NewObject(Place),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return Place, nil
			},
		},
	},
}
