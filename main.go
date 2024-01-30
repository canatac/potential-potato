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

var serverPort = os.Getenv("SERVER_PORT")

var ctx = context.Background()

func connectToRedis() *redis.Client {
	redisURI := os.Getenv("REDIS_URI")
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
		decoder := json.NewDecoder(r.Body)
		var requestBody RequestBody
		err := decoder.Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		otp := requestBody.OTP

		redisClient := connectToRedis()
		if verifyOTP(redisClient, email, otp) {
			redisClient.Del(ctx, email)
			fmt.Fprint(w, "OTP verified")
		} else {
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
