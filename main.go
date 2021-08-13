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
	"golang.org/x/crypto/bcrypt"
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
	ACCESS_SECRET  string `yaml:"secret_key"`
	REFRESH_SECRET string `yaml:"secret_refresh_key"`
}

var pgsql *sqlx.DB
var config = getConfig()
var client *redis.Client
var ctx = context.Background()

func initRedis(ctx context.Context) (*redis.Client, error) {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, err
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

func initDb() (*sqlx.DB, error) {

	db_config := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=%s",
		config.DB.User, config.DB.Password, config.DB.Database, config.DB.Ssl,
	)

	db, err := sqlx.Connect("postgres", db_config)

	if err != nil {
		return nil, err
	}

	return db, nil
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
					return nil, err
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
		"userLogin": &graphql.Field{
			Type: user.LoginMutationType,
			Args: graphql.FieldConfigArgument{
				"login": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				args := params.Args
				var user user.UserLogin

				err := pgsql.Get(&user, "SELECT id, password FROM users WHERE login=$1", args["login"])
				if err != nil {
					return nil, err
				}

				isIncorrectPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(args["password"].(string)))

				if isIncorrectPassword != nil {
					return nil, fmt.Errorf("authorization failed")
				}

				token, err := authorization.CreateToken(user.Id, client, params.Context)

				if err != nil {
					return nil, err
				}

				return map[string]interface{}{"access_token": token.AccessToken, "refresh_token": token.RefreshToken}, nil
			},
		},
		"userCreate": &graphql.Field{
			Type: user.CreateMutationType,
			Args: graphql.FieldConfigArgument{
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
			},

			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				args := params.Args

				var Input user.UserInput

				if val, ok := args["first_name"]; ok {
					Input.First_name = val.(string)
				}

				if val, ok := args["login"]; ok {
					Input.Login = val.(string)
				}

				if val, ok := args["password"]; ok {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(val.(string)), bcrypt.DefaultCost)
					if err != nil {
						return nil, err
					}
					Input.Password = string(hashedPassword)
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

				insertUser := `INSERT INTO users (first_name, last_name, login, password, phone, email) 
				VALUES(:first_name, :last_name, :login, :password, :phone, :email) RETURNING id`
				result, err := pgsql.NamedQuery(insertUser, Input)

				if err != nil {
					return nil, err
				}

				var lastId int
				result.Scan(&lastId)

				token, err := authorization.CreateToken(lastId, client, params.Context)

				if err != nil {
					return nil, fmt.Errorf("unable to create token: %s", err)
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

				userId, err := authorization.FetchAuth(token, client, params.Context)
				if err != nil {
					return nil, fmt.Errorf("undefined user_id")
				}

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
				updateUser := fmt.Sprintf("UPDATE users SET %s WHERE id=$1 RETURNING id", queryString)
				queryResult := pgsql.QueryRowx(updateUser, userId)
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

				token, err := authorization.Refresh(authToken, client, params.Context)
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
			var err error
			client, err = initRedis(r.Context())
			if err != nil {
				log.Fatal(err)
			}
			defer client.Close()
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
	var err error
	pgsql, err = initDb()
	if err != nil {
		log.Fatalf("Connection to the database refused: %s", err)
	}
	defer pgsql.Close()

	os.Setenv("ACCESS_SECRET", config.ACCESS_SECRET)
	os.Setenv("REFRESH_SECRET", config.REFRESH_SECRET)

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
