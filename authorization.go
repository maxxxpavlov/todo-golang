package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var privateKey = getPrivateKey()
var publicKey = getPublicKey()

type JWTClaim struct {
	Username string `json:"username"`

	jwt.StandardClaims
}

func getPrivateKey() rsa.PrivateKey {
	key, err := os.ReadFile("./keys/private.key")
	if err != nil {
		panic(err)
	}
	rsaPrivatePEM, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		panic(err)
	}
	return *rsaPrivatePEM
}
func getPublicKey() rsa.PublicKey {
	key, err := os.ReadFile("./keys/public.key")
	if err != nil {
		panic(err)
	}
	rsaPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		panic(err)
	}
	return *rsaPublicKey
}
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var req, user User
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = DB.Database("test").Collection("users").FindOne(ctx, bson.D{{"username", req.Username}}).Decode(&user)
		if err != nil {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		token, err := generateJWT(req.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := LoginResponse{JWT: token}
		err = respondWithJson(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var req User
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := generateJWT(req.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		newUser := User{Username: req.Username, Password: string(hashedBytes)}
		result, err := DB.Database("test").Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Print(result)
		data := LoginResponse{JWT: token}
		err = respondWithJson(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	tokenString, err := token.SignedString(&privateKey)
	log.Print(err)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return &publicKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaim)

	if !ok {
		err = errors.New("couldn't parse claims")
		return nil, err
	}

	return claims, nil
}
