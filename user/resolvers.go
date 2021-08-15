package user

import (
	"fmt"
	"strings"

	"github.com/Moranilt/go-graphql-location/authorization"
	"github.com/go-redis/redis/v8"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Resolverers interface {
	// Returning a list of all users from table users
	AllUsers() ([]UserType, error)

	// Get information of a single user.
	// It takes user_id from access_token
	User(params graphql.ResolveParams) (*UserType, error)

	// Login action.
	// It takes login and password from graphql.ResolveParams, looking for user by login
	// from table users, comparing hashed password from db with not hashed password from args.
	Login(params graphql.ResolveParams) (*authorization.Tokens, error)

	// Registration of new user.
	// Taking arguments from graphql.ResolveParams and storing it in table "users".
	// Returning access_token and refresh_token for login.
	Create(params graphql.ResolveParams) (*authorization.Tokens, error)

	// Update user ingormation.
	// It takes arguments from graphql.ResolvePrams, authorization key from Context
	// Extracts userId from token and stroing new data to users table.
	Update(params graphql.ResolveParams) (interface{}, error)

	// Action for logout user.
	// It takes authorization header key from graphql.ResolvePrams Context and extracting
	// values from token. Then deleting token from redis client and returning id of user
	Logout(params graphql.ResolveParams) (interface{}, error)

	// Refreshing tokens.
	// It takes refresh token from request and using authorization.RefreshToken
	RefreshToken(params graphql.ResolveParams) (*authorization.Tokens, error)
}

type Resolvers struct {
	pgsql       *sqlx.DB
	RedisClient *redis.Client
}

func GetResolvers(pgsql *sqlx.DB, client *redis.Client) Resolverers {
	return &Resolvers{pgsql: pgsql, RedisClient: client}
}

func (r *Resolvers) User(params graphql.ResolveParams) (*UserType, error) {
	requestToken := params.Context.Value(authorization.AuthHeaderKey)
	token, err := authorization.ExtractTokenMetadata(requestToken.(string))

	if err != nil {
		return nil, err
	}
	userId, err := authorization.FetchAuth(token, r.RedisClient, params.Context)

	if err != nil {
		return nil, err
	}

	var user UserType
	err = r.pgsql.Get(&user, "SELECT * FROM users WHERE id=$1", userId)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Resolvers) AllUsers() ([]UserType, error) {
	var users []UserType

	err := r.pgsql.Select(&users, "SELECT * FROM users")

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Resolvers) Login(params graphql.ResolveParams) (*authorization.Tokens, error) {
	args := params.Args
	var user UserLogin

	err := r.pgsql.Get(&user, "SELECT id, password FROM users WHERE login=$1", args["login"])
	if err != nil {
		return nil, err
	}

	isIncorrectPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(args["password"].(string)))

	if isIncorrectPassword != nil {
		return nil, fmt.Errorf("authorization failed")
	}

	token, err := authorization.CreateToken(user.Id, r.RedisClient, params.Context)

	if err != nil {
		return nil, err
	}

	return &authorization.Tokens{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}, nil
}

func (r *Resolvers) Create(params graphql.ResolveParams) (*authorization.Tokens, error) {
	args := params.Args

	var Input UserInput

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
	result, err := r.pgsql.NamedQuery(insertUser, Input)

	if err != nil {
		return nil, err
	}

	var lastId int
	result.Scan(&lastId)

	token, err := authorization.CreateToken(lastId, r.RedisClient, params.Context)

	if err != nil {
		return nil, fmt.Errorf("unable to create token: %s", err)
	}

	return &authorization.Tokens{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}, nil
}

func (r *Resolvers) Update(params graphql.ResolveParams) (interface{}, error) {
	authToken := params.Context.Value(authorization.AuthHeaderKey).(string)
	token, err := authorization.ExtractTokenMetadata(authToken)

	if err != nil {
		return nil, fmt.Errorf("auth required")
	}

	userId, err := authorization.FetchAuth(token, r.RedisClient, params.Context)
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
	queryResult := r.pgsql.QueryRowx(updateUser, userId)

	var resultUserId int
	queryResult.Scan(&resultUserId)

	return map[string]int{"id": resultUserId}, nil
}

func (r *Resolvers) Logout(params graphql.ResolveParams) (interface{}, error) {
	authToken := params.Context.Value(authorization.AuthHeaderKey).(string)
	token, err := authorization.ExtractTokenMetadata(authToken)

	if err != nil {
		return nil, err
	}

	deleted, delErr := authorization.DeleteAuthFromRedis(token.AccessUuid, r.RedisClient, params.Context)

	if delErr != nil && deleted == 0 {
		return nil, delErr
	}

	return map[string]int{"id": int(deleted)}, nil
}

func (r *Resolvers) RefreshToken(params graphql.ResolveParams) (*authorization.Tokens, error) {
	authToken := params.Args["refresh_token"].(string)
	if authToken == "" {
		return nil, fmt.Errorf("refresh token required")
	}

	token, err := authorization.Refresh(authToken, r.RedisClient, params.Context)
	if err != nil {
		return nil, err
	}

	return &authorization.Tokens{AccessToken: token["access_token"], RefreshToken: token["refresh_token"]}, nil
}
