package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	ACCESS_SECRET  string `yaml:"secret_key"`
	REFRESH_SECRET string `yaml:"secret_refresh_key"`
	RedisDSN       string `yaml:"redis_dsn"`
}

type Repository struct {
	Pgsql         *sqlx.DB
	RedisClient   *redis.Client
	UserResolvers user.Resolverers
}

var config *Config
var repository *Repository

func initRedis(ctx context.Context, redisDSN string) (*redis.Client, error) {
	//Initializing redis
	dsn := redisDSN
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
			Type: graphql.NewList(user.GetTypes().User),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				users, err := repository.UserResolvers.AllUsers()

				if err != nil {
					return nil, err
				}

				return users, nil
			},
		},
		"user": &graphql.Field{
			Type: user.GetTypes().User,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user, err := repository.UserResolvers.User(params)

				if err != nil {
					return nil, err
				}

				return user, nil
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
	Name: "Mutations",
	Fields: graphql.Fields{
		"userLogin": &graphql.Field{
			Type: user.GetTypes().Login,
			Args: user.GetArguments().Login,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				tokens, err := repository.UserResolvers.Login(params)

				if err != nil {
					return nil, err
				}

				return tokens, nil
			},
		},
		"userCreate": &graphql.Field{
			Type: user.GetTypes().Create,
			Args: user.GetArguments().Create,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				tokens, err := repository.UserResolvers.Create(params)

				if err != nil {
					return nil, err
				}

				return tokens, nil
			},
		},
		"userUpdate": &graphql.Field{
			Type: user.GetTypes().Update,
			Args: user.GetArguments().Update,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				result, err := repository.UserResolvers.Update(params)

				if err != nil {
					return nil, err
				}

				return result, nil
			},
		},
		"userLogout": &graphql.Field{
			Type: user.GetTypes().Logout,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				result, err := repository.UserResolvers.Logout(params)

				if err != nil {
					return nil, err
				}

				return result, nil
			},
		},
		"userRefreshToken": &graphql.Field{
			Type: user.GetTypes().RefreshToken,
			Args: user.GetArguments().RefreshToken,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				tokens, err := repository.UserResolvers.RefreshToken(params)

				if err != nil {
					return nil, err
				}

				return tokens, nil
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
	var globalContext = context.Background()

	config = getConfig()

	pgsql, err := initDb()
	if err != nil {
		log.Fatalf("Connection to the database refused: %s", err)
	}
	defer pgsql.Close()

	redisClient, err := initRedis(globalContext, config.RedisDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	os.Setenv("ACCESS_SECRET", config.ACCESS_SECRET)
	os.Setenv("REFRESH_SECRET", config.REFRESH_SECRET)

	repository = &Repository{
		RedisClient:   redisClient,
		Pgsql:         pgsql,
		UserResolvers: user.GetResolvers(pgsql, redisClient),
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
