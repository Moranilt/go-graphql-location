package payments

import "github.com/graphql-go/graphql"

type Arguments struct {
	Create graphql.FieldConfigArgument
}

func GetArguments() Arguments {
	return Arguments{
		Create: createArgs,
	}
}

var createArgs = graphql.FieldConfigArgument{
	"amount": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.Float),
	},
}
