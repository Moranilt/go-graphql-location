package payments

import (
	"fmt"

	"github.com/Moranilt/go-graphql-location/authorization"
	"github.com/go-redis/redis/v8"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

type Resolverers interface {
	Create(graphql.ResolveParams) (*PaymentsCreateReturn, error)
}

type PaymentsCreateReturn struct {
	Id int `json:"id" db:"id"`
}

type PaymentsInput struct {
	UserId uint64  `json:"user_id" db:"user_id"`
	Amount float64 `json:"amount" db:"amount"`
}

type Resolvers struct {
	pgsql       *sqlx.DB
	RedisClient *redis.Client
}

func GetResolvers(pgsql *sqlx.DB, client *redis.Client) *Resolvers {
	return &Resolvers{pgsql: pgsql, RedisClient: client}
}

func (r *Resolvers) Create(params graphql.ResolveParams) (*PaymentsCreateReturn, error) {
	requestToken := params.Context.Value(authorization.AuthHeaderKey)
	token, err := authorization.ExtractTokenMetadata(requestToken.(string))
	args := params.Args
	var PaymentInput PaymentsInput

	if err != nil {
		return nil, err
	}

	userId, err := authorization.FetchAuth(token, r.RedisClient, params.Context)

	PaymentInput.UserId = userId

	if err != nil {
		return nil, err
	}

	if val, ok := args["amount"]; ok {
		PaymentInput.Amount = val.(float64)
	}

	if PaymentInput.Amount == 0 {
		return nil, fmt.Errorf("amount should not be empty")
	}

	insertPayment := `INSERT INTO payments (user_id, amount) VALUES(:user_id, :amount) RETURNING id`

	result, err := r.pgsql.NamedQuery(insertPayment, PaymentInput)

	if err != nil {
		return nil, err
	}

	var paymentResult PaymentsCreateReturn

	if result.Next() {
		result.Scan(&paymentResult.Id)
	}

	return &paymentResult, nil
}
