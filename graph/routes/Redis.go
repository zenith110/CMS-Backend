package routes

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

type UserData struct {
	Password string `json:"password"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

func RedisClientInstation() *redis.Client {
	redisclient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDISURL"),
		Password: os.Getenv("REDISPASSWORD"),
		DB:       0, // use default DB
	})

	return redisclient
}
func MarshalBinary(userData map[string]string) ([]byte, error) {
	return json.Marshal(userData)
}
func RedisUserInfo(jwt string, redisClient *redis.Client) map[string]string {
	data, err := redisClient.Get(jwt).Result()
	switch {
	case err == redis.Nil:
		fmt.Println("key does not exist")
	case err != nil:
		fmt.Println("Get failed", err)
	case data == "":
		fmt.Println("value is empty")
	}
	var user UserData
	marshalErr := json.Unmarshal([]byte(data), &user)
	if marshalErr != nil {
		panic(fmt.Sprintf("Error has occured while trying to grab user data! %v\n", marshalErr))
	}

	username := user.Username
	password := user.Password
	role := user.Role
	redisData := map[string]string{
		"username": username,
		"password": password,
		"role":     role,
	}
	return redisData
}
