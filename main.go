package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Moranilt/go-graphql-location/authorization"
	"github.com/Moranilt/go-graphql-location/objects"
	"github.com/Moranilt/go-graphql-location/user"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
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
	DB             DBConfig
	Secret         string `yaml:"secret_key"`
	Secret_refresh string `yaml:"secret_refresh_key"`
}

var pgsql *sqlx.DB
var config = getConfig()
var client *redis.Client
var ctx = context.Background()

func initRedis() {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	client = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

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
		"userCreate": &graphql.Field{
			Type: user.CreateMutationType,
			Args: graphql.FieldConfigArgument{
				"first_name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"last_name": &graphql.ArgumentConfig{
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

				if val, ok := args["first_name"]; ok {
					Input.First_name = val.(string)
				}

				if val, ok := args["last_name"]; ok {
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

				token, err := authorization.CreateToken(lastId)

				if err != nil {
					log.Fatalf("Unable to create token: %s", err)
				}

				saveErr := authorization.CreateRedisAuth(uint64(lastId), token, client, ctx)

				if saveErr != nil {
					log.Fatalf("Error while adding redis key: %s", saveErr)
				}

				return map[string]interface{}{"id": lastId, "access_token": token.AccessToken, "refresh_token": token.RefreshToken}, nil
			},
		},
		"userUpdate": &graphql.Field{
			Type: user.UpdateMutationType,
			Args: graphql.FieldConfigArgument{
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
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				authToken := params.Context.Value(authorization.AuthHeaderKey).(string)
				token, err := authorization.ExtractTokenMetadata(authToken)

				if err != nil {
					return nil, fmt.Errorf("auth required")
				}

				userId, err := authorization.FetchAuth(token, client, ctx)
				if err != nil {
					return nil, fmt.Errorf("undefined user_id")
				}
				fmt.Println(userId)

				updateColumns := []string{
					"first_name",
					"last_name",
					"phone",
					"email",
				}

				var result []string

				for _, columnName := range updateColumns {
					if val, ok := params.Args[columnName]; ok {
						result = append(result, columnName+"='"+val.(string)+"'")
					}
				}

				queryString := strings.Join(result, ",")
				fmt.Println(queryString)
				updateUser := fmt.Sprintf("UPDATE users SET %s WHERE id=$1 RETURNING id", queryString)
				queryResult := pgsql.QueryRowx(updateUser, userId)
				fmt.Println(queryResult.Err())
				var resultUserId int
				queryResult.Scan(&resultUserId)

				return map[string]int{"id": resultUserId}, nil
			},
		},
		"userLogout": &graphql.Field{
			Type: user.LogoutMutationType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				authToken := params.Context.Value(authorization.AuthHeaderKey).(string)
				token, err := authorization.ExtractTokenMetadata(authToken)

				if err != nil {
					return nil, err
				}

				deleted, delErr := authorization.DeleteAuthFromRedis(token.AccessUuid, client, ctx)

				if delErr != nil && deleted == 0 {
					return nil, delErr
				}

				return map[string]int{"id": int(deleted)}, nil
			},
		},
		"userRefreshToken": &graphql.Field{
			Type: user.RefreshTokenMutationType,
			Args: graphql.FieldConfigArgument{
				"refresh_token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				authToken := params.Args["refresh_token"].(string)
				if authToken == "" {
					return nil, fmt.Errorf("refresh token required")
				}

				token, err := authorization.Refresh(authToken, client, ctx)
				fmt.Println(err)
				if err != nil {
					return nil, err
				}

				return map[string]string{"access_token": token["access_token"], "refresh_token": token["refresh_token"]}, nil
			},
		},
	},
})

func customHandler(schema *graphql.Schema) http.Handler {
	r := mux.NewRouter()
	r.Path("/").Handler(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// store authorization header to context
			next.ServeHTTP(
				w,
				r.WithContext(
					context.WithValue(r.Context(),
						authorization.AuthHeaderKey,
						authorization.GetAuthToken(r)),
				),
			)
		})
	}(
		gqlhandler.New(&gqlhandler.Config{
			Schema:     schema,
			Pretty:     true,
			Playground: true,
			GraphiQL:   false,
		})))

	return r
}

func main() {
	err := initDb()
	initRedis()
	os.Setenv("ACCESS_SECRET", config.Secret)
	os.Setenv("REFRESH_SECRET", config.Secret_refresh)

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

	mux := http.NewServeMux()
	mux.Handle("/", customHandler(&schema))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
