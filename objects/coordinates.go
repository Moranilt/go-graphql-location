package objects

import "github.com/graphql-go/graphql"

var Coordinates = graphql.ObjectConfig{
	Name: "Coordinates",
	Fields: graphql.Fields{
		"x": &graphql.Field{
			Type: graphql.Float,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return 23.5, nil
			},
		},
		"y": &graphql.Field{
			Type: graphql.Float,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return 58.4, nil
			},
		},
	},
}
