package user

import "github.com/graphql-go/graphql"

var LoginArgs = graphql.FieldConfigArgument{
	"login": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"password": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}

var CreateArgs = graphql.FieldConfigArgument{
	"first_name": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"last_name": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"login": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"password": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"phone": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"email": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}

var UpdateArgs = graphql.FieldConfigArgument{
	"first_name": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"last_name": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"phone": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"email": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var RefreshTokenArgs = graphql.FieldConfigArgument{
	"refresh_token": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}
