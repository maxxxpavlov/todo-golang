package main

import (
	"context"
	"crypto/rsa"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
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

var privateKey = getPrivateKey()

var DB = getDBConnection()

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(30 * time.Minute)
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
func getPrivateKey() rsa.PrivateKey {
	key, err := os.ReadFile("./private.key")
	if err != nil {
		panic(err)
	}
	rsaPrivatePEM, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		panic(err)
	}
	return *rsaPrivatePEM
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", index)
	mux.HandleFunc("/user/login", login)
	mux.HandleFunc("/user/register", register)
	c := cors.New(cors.Options{AllowedOrigins: []string{"*"}})
	handler := c.Handler(mux)
	s := &http.Server{Addr: ":3000", Handler: handler}
	log.Fatal(s.ListenAndServe())
}
