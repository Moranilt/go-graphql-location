package user

import "github.com/graphql-go/graphql"

type Arguments struct {
	Login        graphql.FieldConfigArgument
	Create       graphql.FieldConfigArgument
	Update       graphql.FieldConfigArgument
	RefreshToken graphql.FieldConfigArgument
}

func GetArguments() Arguments {
	return Arguments{
		Login:        loginArgs,
		Create:       createArgs,
		Update:       updateArgs,
		RefreshToken: refreshTokenArgs,
	}
}

var loginArgs = graphql.FieldConfigArgument{
	"login": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"password": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}

var createArgs = graphql.FieldConfigArgument{
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

var updateArgs = graphql.FieldConfigArgument{
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

var refreshTokenArgs = graphql.FieldConfigArgument{
	"refresh_token": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}
