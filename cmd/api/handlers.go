package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/basedalex/medods-test/data"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	UserID string `json:"user_id"`
}

type RefreshRequest struct {
	UserID string `json:"user_id"`
	AuthToken string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AuthToken string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}

func (app *Config) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
}

func (app *Config) Refresh(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	var req RefreshRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.AuthToken == "" || req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secretString := []byte("supersecretstring")

	authToken, err := jwt.Parse(req.AuthToken, func(token *jwt.Token) (interface{}, error) {
		return secretString, nil
	})
	if err != nil {
		log.Println("could not parse token", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims, ok := authToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("unexpected token claims %T: %s", authToken.Claims, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("refresh_token")

	var entry data.RefreshToken
	
	fmt.Println(claims)

	err = collection.FindOne(ctx, bson.M{"userid": claims["userID"]}).Decode(&entry)
	fmt.Println()
	if err != nil {
		log.Println("error getting", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestRefreshToken, err := base64.StdEncoding.DecodeString(req.RefreshToken)
	if err != nil {
		log.Println("error decoding refresh token",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(entry.Token), requestRefreshToken)
	if err != nil {
		log.Println("wrong refresh token", err)
		log.Println(entry.Token, string(requestRefreshToken))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqUserID, ok := claims["userID"].(string)
	if !ok {
		log.Println("Not a string", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newAuthToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"userID": reqUserID,
	})

	authTokenString, err := newAuthToken.SignedString(secretString)
	if err != nil {
		log.Println("Error creating token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshTokenString := generateRandomString(20)

	bcryptedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenString), 10)
	if err != nil {
		log.Println("Error creating token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	base64RefreshToken := base64.StdEncoding.EncodeToString([]byte(refreshTokenString))


	fmt.Println("user_id is:", claims["userID"] )
	fmt.Println("bcrypted refresh token", string(bcryptedRefreshToken))

	
	filter := bson.D{{"userid", reqUserID}}

	update := bson.D{{"$set", bson.D{{"token", string(bcryptedRefreshToken)}}}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		log.Println("Error inserting", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		AuthToken: authTokenString,
		RefreshToken: base64RefreshToken,
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Panicln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}


func (app *Config) Auth(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	var req AuthRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		log.Println("no user_id is specified")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"userID": req.UserID,
	})

	collection := client.Database("logs").Collection("refresh_token")

	secretString := []byte("supersecretstring")

	authTokenString, err := authToken.SignedString(secretString)
	if err != nil {
		log.Println("Error creating token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshTokenString := generateRandomString(20)

	bcryptedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenString), 10)
	if err != nil {
		log.Println("Error creating token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	base64RefreshToken := base64.StdEncoding.EncodeToString([]byte(refreshTokenString))

	_, err = collection.InsertOne(context.TODO(), data.RefreshToken{
		Token: string(bcryptedRefreshToken),
		UserID: req.UserID,
	})


	if err != nil {
		log.Println("Error inserting", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		AuthToken: authTokenString,
		RefreshToken: base64RefreshToken,
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Panicln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
