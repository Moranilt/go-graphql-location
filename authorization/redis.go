package authorization

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
)

func FetchAuth(authD *AccessDetails, client *redis.Client, ctx context.Context) (uint64, error) {
	redisUserId, err := client.Get(ctx, authD.AccessUuid).Result()

	if err != nil {
		return 0, err
	}

	userId, _ := strconv.ParseUint(redisUserId, 10, 64)

	return userId, nil
}

func createRedisAuth(userid uint64, td *TokenDetails, client *redis.Client, ctx context.Context) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := client.Set(ctx, td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	errRefresh := client.Set(ctx, td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func DeleteAuthFromRedis(givenUuid string, client *redis.Client, ctx context.Context) (int64, error) {
	deleted, err := client.Del(ctx, givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func Refresh(refresh_token string, client *redis.Client, ctx context.Context) (map[string]string, error) {
	token, err := jwt.Parse(refresh_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string)

		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)

		if err != nil {
			return nil, err
		}

		isContainsInRedis := client.Get(ctx, refreshUuid)

		if isContainsInRedis.Err() != nil {
			return nil, fmt.Errorf("refresh token is expired")
		}

		deleted, delErr := DeleteAuthFromRedis(refreshUuid, client, ctx)

		if delErr != nil && deleted == 0 {
			return nil, delErr
		}

		ts, createErr := CreateToken(int(userId), client, ctx)

		if createErr != nil {
			return nil, createErr
		}

		saveErr := createRedisAuth(userId, ts, client, ctx)

		if saveErr != nil {
			return nil, saveErr
		}

		return map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}, nil

	}

	return nil, fmt.Errorf("unexpected token")
}
