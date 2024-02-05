package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	redis "github.com/go-redis/redis/v8"
)

var redisURI = os.Getenv("REDIS_URI")

var ctx = context.Background()

func connectToRedis() *redis.Client {
	if redisURI == "" {
		panic("REDIS_URI is empty")
	}

	addr, err := redis.ParseURL(redisURI)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(addr)

	_, err = client.Ping(ctx).Result()

	if err != nil {
		panic(err)
	}

	return client
}

func verifyOTP(redisClient *redis.Client, email, otp string) bool {
	storedOTP, err := redisClient.Get(ctx, email).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("OTP not found in Redis for email:", email)
			return false
		}
		log.Println("Error retrieving OTP from Redis:", err)
		return false
	}

	return otp == storedOTP
}

type RequestBody struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func main() {
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestBody RequestBody

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		otp := requestBody.OTP

		redisClient := connectToRedis()
		if verifyOTP(redisClient, email, otp) {
			redisClient.Del(ctx, email)
			fmt.Fprint(w, "OTP verified")
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
