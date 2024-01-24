package main

import (
	"context"
	"fmt"
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
		return false
	}

	return otp == storedOTP
}

func main() {
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		otp := r.URL.Query().Get("otp")

		redisClient := connectToRedis()
		if verifyOTP(redisClient, email, otp) {
			redisClient.Del(ctx, email)
			fmt.Fprint(w, "OTP verified")
		} else {
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%s", serverPort, nil))
}
