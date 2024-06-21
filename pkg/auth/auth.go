package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"

	"github.com/golang-jwt/jwt"
)

const (
	_accessTokenLifetime = time.Hour
)

var jwtPrivateKey *rsa.PrivateKey

var jwtMiddleware *jwtmiddleware.JWTMiddleware

func InitKeys() error {
	jwtPublicBytes, err := os.ReadFile("./public.pem")
	if err != nil {
		return err
	}

	jwtPrivateBytes, err := os.ReadFile("./private.pem")
	if err != nil {
		return err
	}

	jwtPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(jwtPublicBytes)
	if err != nil {
		return err
	}

	jwtPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(jwtPrivateBytes)
	if err != nil {
		return err
	}

	keyFunc := func(ctx context.Context) (interface{}, error) {
		return jwtPublicKey, nil
	}

	jwtValidator, err := validator.New(
		keyFunc,
		validator.RS256,
		"https://akruglov.ru",
		[]string{"kruglov"},
	)
	if err != nil {
		return err
	}

	jwtMiddleware = jwtmiddleware.New(jwtValidator.ValidateToken)

	return nil
}

func CreateAccessToken(userId int64) string {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)

	claims["id"] = userId
	claims["exp"] = time.Now().Add(_accessTokenLifetime).Unix()

	token.Claims = claims

	tokenString, _ := token.SignedString(jwtPrivateKey)

	return tokenString
}

func GetHandlerWithJwt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtMiddleware.CheckJWT(h).ServeHTTP(w, r)
	})
}

func CreateRefreshToken() string {
	b := make([]byte, 46)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GetUserIdByToken(r *http.Request) string {
	return ""
}
