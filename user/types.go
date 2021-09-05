package user

import (
	"github.com/Moranilt/go-graphql-location/payments"
	"github.com/graphql-go/graphql"
)

// inheritance is not working for structs with graphql.ResolverFunc
type UserType struct {
	Id         int                `json:"_id"`
	Created_at string             `json:"created_at"`
	Updated_at string             `json:"updated_at"`
	First_name string             `json:"first_name" db:"first_name"`
	Last_name  string             `json:"last_name" db:"last_name"`
	Phone      string             `json:"phone" db:"phone"`
	Email      string             `json:"email" db:"email"`
	Login      string             `json:"login" db:"login"`
	Password   string             `json:"password" db:"password"`
	Last_login string             `json:"last_login" db:"last_login"`
	Payments   []payments.Payment `json:"payments"`
	// UserInput
}

// struct for storing users data
type UserInput struct {
	First_name string `json:"first_name" db:"first_name"`
	Last_name  string `json:"last_name" db:"last_name"`
	Phone      string `json:"phone" db:"phone"`
	Email      string `json:"email" db:"email"`
	Login      string `json:"login" db:"login"`
	Password   string `json:"password" db:"password"`
}

type UserLogin struct {
	Id       int    `json:"id" db:"id"`
	Password string `json:"password" db:"password"`
}

type Types struct {
	User         *graphql.Object
	Create       *graphql.Object
	Login        *graphql.Object
	Update       *graphql.Object
	Logout       *graphql.Object
	RefreshToken *graphql.Object
}

func GetTypes() Types {
	return Types{
		User:         userType,
		Create:       createType,
		Login:        loginType,
		Update:       updateType,
		Logout:       logoutType,
		RefreshToken: refreshTokenType,
	}
}

var userType = graphql.NewObject(graphql.ObjectConfig{
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
			"last_login": &graphql.Field{
				Type:        graphql.String,
				Description: "Users last login",
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
			"payments": &graphql.Field{
				Type:        graphql.NewList(paymentType),
				Description: "Users Payments",
			},
		}
	}),
})

var paymentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PaymentType",
	Fields: graphql.Fields{
		"user_id": &graphql.Field{
			Type: graphql.Int,
		},
		"amount": &graphql.Field{
			Type: graphql.Float,
		},
		"payed": &graphql.Field{
			Type: graphql.Boolean,
		},
		"payed_at": &graphql.Field{
			Type: graphql.String,
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var createType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateUser",
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

var loginType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthUser",
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

var updateType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateUser",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})

var logoutType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LogoutUser",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		}
	}),
})

var refreshTokenType = graphql.NewObject(graphql.ObjectConfig{
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
