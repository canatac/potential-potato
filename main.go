package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	redis "github.com/go-redis/redis/v8"
)

var dbUrl = os.Getenv("DB_URL")
var dbPort = os.Getenv("DB_PORT")
var serverPort = os.Getenv("SERVER_PORT")

var ctx = context.Background()

func connectToRedis(redisAddr, redisPassword string) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword, // no password set
        DB:       0,  // use default DB
    })

    _, err := client.Ping(ctx).Result()
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

func main(){
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		otp := r.URL.Query().Get("otp")
	
		redisClient := connectToRedis(fmt.Sprintf("%s:%s",dbUrl,dbPort), "")
		if verifyOTP(redisClient, email, otp) {
			redisClient.Del(ctx, email)
			fmt.Fprint(w, "OTP verified")
		} else {
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		}
	})
	
	http.ListenAndServe(fmt.Sprintf(":%s",serverPort, nil))
}
