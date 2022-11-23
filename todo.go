package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GETTODORequest struct {
	Page uint `json:"page"`
}
type POSTTODORequest struct {
	Text string `json:"text"`
}
type PUTTODORequest struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id, omitempty"`
	Text string             `json:"text"`
}
type TODO struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id, omitempty"`
	Creator string             `json:"creator"`
	Text    string             `json:"text"`
}
type TODOResponse struct {
	Todos []TODO `json:"todos"`
}
type TokenBody struct {
	Token string `json:"token"`
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	var token TokenBody
	if r.Method == http.MethodGet {
		token.Token = r.URL.Query().Get("token")
	} else {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&token)
		if err != nil {
			http.Error(w, "No token", http.StatusUnauthorized)
			return
		}
	}
	claims, err := ValidateToken(token.Token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	username := claims.Username
	switch r.Method {
	case http.MethodGet:
		getTODO(w, r)
	case http.MethodPost:
		createTODO(w, r, username)
	case http.MethodPut:
		updateTODO(w, r)
	case http.MethodDelete:
		deleteTODO(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func getTODO(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var todos []TODO
	todos = make([]TODO, 0)

	results, err := DB.Database("test").Collection("todos").Find(ctx, bson.D{{}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for results.Next(ctx) {
		var todo TODO

		if err = results.Decode(&todo); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	respondWithJson(w, TODOResponse{todos})

}
func createTODO(w http.ResponseWriter, r *http.Request, username string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var todo POSTTODORequest
	log.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Print(err)

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newTodo := TODO{Text: todo.Text, Creator: username}
	newSavedTODOResult, err := DB.Database("test").Collection("todos").InsertOne(ctx, bson.M{"text": newTodo.Text, "creator": newTodo.Creator})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newTodo.ID = newSavedTODOResult.InsertedID.(primitive.ObjectID)

	respondWithJson(w, newTodo)

}
func updateTODO(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var todo PUTTODORequest
	err := json.NewDecoder(r.Body).Decode(&todo)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = DB.Database("test").Collection("todos").UpdateByID(ctx, todo.ID, bson.M{"$set": bson.M{"text": todo.Text}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func deleteTODO(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var todo PUTTODORequest
	err := json.NewDecoder(r.Body).Decode(&todo)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = DB.Database("test").Collection("todos").DeleteOne(ctx, bson.M{"_id": todo.ID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
