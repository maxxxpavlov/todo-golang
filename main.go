package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LoginResponse struct {
	JWT string `json:"jwt"`
}
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var DB = getDBConnection()

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hi man")
}

func getDBConnection() *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongourl, exists := os.LookupEnv("mongourl")
	var url = "mongodb://root:example@localhost:27017"
	if exists {
		url = mongourl
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	return client
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", index)
	mux.HandleFunc("/user/login", login)
	mux.HandleFunc("/user/register", register)
	mux.HandleFunc("/todo", todoHandler)
	c := cors.New(cors.Options{AllowedOrigins: []string{"*"}})
	handler := c.Handler(mux)
	s := &http.Server{Addr: ":3000", Handler: handler}
	log.Fatal(s.ListenAndServe())
}
