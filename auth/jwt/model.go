package jwt

import (
	"errors"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type JWTBody struct {
	Id    string
	Email string
}

var (
	ErrInvalidID          = errors.New("invalid ID in jwt body")
	ErrInvalidEmail       = errors.New("invalid Email in jwt body")
	ErrMissingLoginValues = jwt.ErrMissingLoginValues
)

func BodyFromContext(c *gin.Context) (*JWTBody, error) {
	claims := jwt.ExtractClaims(c)
	id, ok := claims["Id"].(string)
	if !ok {
		return nil, ErrInvalidID
	}
	email, ok := claims["Email"].(string)
	if !ok {
		return nil, ErrInvalidEmail
	}

	return &JWTBody{
		Id:    id,
		Email: email,
	}, nil
}

func GetToken(c *gin.Context) string {
	return jwt.GetToken(c)
}
