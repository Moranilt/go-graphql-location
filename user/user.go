package user

import "github.com/graphql-go/graphql"

type UserType struct {
	Id         int    `json:"_id"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
	UserInput
}
type UserInput struct {
	First_name string `json:"firstname" db:"first_name"`
	Last_name  string `json:"lastname" db:"last_name"`
	Phone      string `json:"phone" db:"phone"`
	Email      string `json:"email" db:"email"`
}

var QueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"first_name": &graphql.Field{
				Type:        graphql.String,
				Description: "Users firstname",
			},
			"last_name": &graphql.Field{
				Type:        graphql.String,
				Description: "Users lastname",
			},
			"phone": &graphql.Field{
				Type:        graphql.String,
				Description: "Phone number",
			},
			"email": &graphql.Field{
				Type:        graphql.String,
				Description: "Email",
			},
			"_id": &graphql.Field{
				Type:        graphql.String,
				Description: "users id",
			},
			"created_at": &graphql.Field{
				Type:        graphql.String,
				Description: "Registration date",
			},
			"updated_at": &graphql.Field{
				Type:        graphql.String,
				Description: "Updated profile date",
			},
		}
	}),
})

var CreateMutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateUser",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"access_token": &graphql.Field{
				Type: graphql.String,
			},
			"refresh_token": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})

var UpdateMutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateUser",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})

var LogoutMutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LogoutUser",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})

var RefreshTokenMutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "RefreshToken",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"access_token": &graphql.Field{
				Type: graphql.String,
			},
			"refresh_token": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})
