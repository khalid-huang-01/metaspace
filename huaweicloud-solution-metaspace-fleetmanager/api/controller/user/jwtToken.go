package user

import (
	"fleetmanager/setting"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Claims struct {
	SessionId string 
	UserId    string 
	jwt.StandardClaims
}

func GetJWTToken(sessionId string, id string, expireTime time.Time) (string, error) {
	claims := &Claims{
		SessionId: sessionId,
		UserId:    id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(setting.JwtKey))
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(setting.JwtKey), nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, errors.New("Invalid token")
}
