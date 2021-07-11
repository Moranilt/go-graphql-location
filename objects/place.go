package objects

import "github.com/graphql-go/graphql"

var Place = graphql.ObjectConfig{
	Name: "Place",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Pizza Gerto", nil
			},
		},
		"phone": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "79824872134", nil
			},
		},
	},
}
