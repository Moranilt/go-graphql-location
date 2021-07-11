package objects

import "github.com/graphql-go/graphql"

var User = graphql.ObjectConfig{
	Name:        "User",
	Description: "Get user info",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "Users name",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Lily", nil
			},
		},
		"age": &graphql.Field{
			Type:        graphql.Int,
			Description: "Users age",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return 22, nil
			},
		},
	},
}
