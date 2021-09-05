package payments

import "github.com/graphql-go/graphql"

type Types struct {
	CreatePayment *graphql.Object
}

type Payment struct {
	Id        uint64  `json:"id" db:"id"`
	UserId    uint64  `json:"user_id" db:"user_id"`
	Amount    float64 `json:"amount" db:"amount"`
	Payed     bool    `json:"payed" db:"payed"`
	CreatedAt string  `json:"created_at" db:"created_at"`
	PayedAt   *string `json:"payed_at" db:"payed_at"`
}

func GetTypes() Types {
	return Types{
		CreatePayment: createPayment,
	}
}

var createPayment = graphql.NewObject(graphql.ObjectConfig{
	Name:        "createPayment",
	Description: "User create payment",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Float,
			},
		}
	}),
})
