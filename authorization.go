package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var req, user User
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = DB.Database("todo").Collection("users").FindOne(ctx, bson.D{{"username", req.Username}}).Decode(&user)
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
	if r.Method == "POST" {
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
		result, err := DB.Database("todo").Collection("users").InsertOne(ctx, newUser)
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
