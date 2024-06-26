package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	_accessTokenLifetime = time.Hour
)

var jwtPrivateKey *rsa.PrivateKey

var jwtMiddleware *jwtmiddleware.JWTMiddleware

type CustomClaims struct {
	UserID int64 `json:"userId"`
}

func (c *CustomClaims) Validate(_ context.Context) error {
	return nil
}

type GophermartClaims struct {
	CustomClaims
	jwt.RegisteredClaims
}

func jwtErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case errors.Is(err, jwtmiddleware.ErrJWTMissing):
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"JWT is missing."}`))
	case errors.Is(err, jwtmiddleware.ErrJWTInvalid):
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"JWT is invalid."}`))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Something went wrong while checking the JWT."}`))
	}
}

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

	customClaims := func() validator.CustomClaims {
		return &CustomClaims{}
	}

	jwtValidator, err := validator.New(
		keyFunc,
		validator.RS256,
		"https://akruglov.ru",
		[]string{"kruglov"},
		validator.WithCustomClaims(customClaims),
	)
	if err != nil {
		return err
	}

	jwtMiddleware = jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(jwtErrorHandler),
	)

	return nil
}

func CreateAccessToken(userID int64) string {
	claims := GophermartClaims{
		CustomClaims{
			UserID: userID,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(_accessTokenLifetime)),
			Issuer:    "https://akruglov.ru",
			Audience:  []string{"kruglov"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, _ := token.SignedString(jwtPrivateKey)

	return "Bearer " + tokenString
}

func GetHandlerWithJwt(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtMiddleware.CheckJWT(h).ServeHTTP(w, r)
	})
}

func CreateRefreshToken() string {
	b := make([]byte, 46)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GetUserIDByToken(r *http.Request) *int64 {
	token := r.Context().Value(jwtmiddleware.ContextKey{})
	if token == nil {
		return nil
	}

	claims := token.(*validator.ValidatedClaims)
	customClaims := claims.CustomClaims.(*CustomClaims)

	return &customClaims.UserID
}
