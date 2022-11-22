package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
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

var privateKey *rsa.PrivateKey

var DB = getDBConnection()

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
	claims["username"] = username
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hi man")
}
func respondWithJson(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

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

func getDBConnection() *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
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
	key, err := os.ReadFile("./private.key")
	if err != nil {
		panic(err)
	}
	rsaPrivatePEM, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		panic(err)
	}
	privateKey = rsaPrivatePEM

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", index)
	mux.HandleFunc("/user/login", login)
	mux.HandleFunc("/user/register", register)
	c := cors.New(cors.Options{AllowedOrigins: []string{"*"}})
	handler := c.Handler(mux)
	s := &http.Server{Addr: ":3000", Handler: handler}
	log.Fatal(s.ListenAndServe())
}
