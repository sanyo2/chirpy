package auth

import (
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashed_password, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	isSame, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return isSame, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	currentTime := time.Now().UTC()
	expirationTime := currentTime.Add(expiresIn)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   userID.String(),
	})

	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	claim := jwt.RegisteredClaims{}
	fmt.Println(tokenString)
	fmt.Println(tokenSecret)
	token, err := jwt.ParseWithClaims(tokenString, &claim,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)

	if err != nil {
		return uuid.Nil, err
	}

	uuidString, err := token.Claims.GetSubject()

	if err != nil {
		return uuid.Nil, err
	}

	uuidValue, err := uuid.Parse(uuidString)

	if err != nil {
		return uuid.Nil, err
	}

	fmt.Println("4")

	return uuidValue, nil
}
