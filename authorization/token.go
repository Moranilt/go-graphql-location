package authorization

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type AuthHeaderKeyType string

const AuthHeaderKey AuthHeaderKeyType = "Authorization"

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

func CreateToken(userId int, client *redis.Client, ctx context.Context) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.New().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.New().String()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["user_id"] = userId
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["authorized"] = true
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))

	if err != nil {
		return nil, err
	}

	err = CreateRedisAuth(uint64(userId), td, client, ctx)

	if err != nil {
		return nil, err
	}

	return td, nil
}

func GetAuthToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	strArr := strings.Split(bearerToken, " ")

	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func VerifyToken(authHeaderToken string) (*jwt.Token, error) {
	tokenString := authHeaderToken
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func TokenValid(r string) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("token is not valid")
	}

	return nil
}

func ExtractTokenMetadata(r string) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)

		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}

		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}

	return nil, err
}
