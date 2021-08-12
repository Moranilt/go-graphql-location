package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Moranilt/go-graphql-location/objects"
	"github.com/Moranilt/go-graphql-location/user"
	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/graphql-go-handler"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

type DBConfig struct {
	User     string `yaml:"USER"`
	Password string `yaml:"PASSWORD"`
	Database string `yaml:"DATABASE"`
	Ssl      string `yaml:"SSL_MODE"`
}
type Config struct {
	DB DBConfig
}

var pgsql *sqlx.DB

func getConfig() *Config {
	content, err := ioutil.ReadFile("./config.yml")

	if err != nil {
		log.Fatal(err)
	}
	config := &Config{}
	err = yaml.Unmarshal(content, &config)

	if err != nil {
		log.Fatal(err)
	}

	return config
}

func initDb() error {
	config := getConfig()

	db_config := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=%s",
		config.DB.User, config.DB.Password, config.DB.Database, config.DB.Ssl,
	)

	db, e := sqlx.Connect("postgres", db_config)

	if e != nil {
		return e
	}

	pgsql = db

	return nil
}

var Query = graphql.NewObject(graphql.ObjectConfig{
	Name: "GeoLocationRemember",
	Fields: graphql.Fields{
		"users": &graphql.Field{
			Type: graphql.NewList(user.QueryType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var users []user.UserType

				err := pgsql.Select(&users, "SELECT * FROM users")

				if err != nil {
					log.Fatal(err)
				}

				return users, nil
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

var Mutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "StoreSomething",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: user.MutationType,
			Args: graphql.FieldConfigArgument{
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"phone": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},

			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				args := params.Args

				var Input user.UserInput

				if val, ok := args["firstname"]; ok {
					Input.First_name = val.(string)
				}

				if val, ok := args["lastname"]; ok {
					Input.Last_name = val.(string)
				}

				if val, ok := args["phone"]; ok {
					Input.Phone = val.(string)
				}

				if val, ok := args["email"]; ok {
					Input.Email = val.(string)
				}

				insertUser := `INSERT INTO users (first_name, last_name, phone, email) VALUES($1, $2, $3, $4) RETURNING id`
				result := pgsql.QueryRowx(insertUser, Input.First_name, Input.Last_name, Input.Phone, Input.Email)

				var lastId int
				result.Scan(&lastId)

				return map[string]int{"id": lastId}, nil
			},
		},
	},
})

func main() {
	err := initDb()

	if err != nil {
		log.Fatalf("Connection to the database refused: %s", err)
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    Query,
		Mutation: Mutation,
	})

	if err != nil {
		log.Fatalf("Failed to create new schema. Error: %v", err)
	}

	handler := gqlhandler.New(&gqlhandler.Config{
		Schema:     &schema,
		Pretty:     true,
		Playground: true,
	})

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
