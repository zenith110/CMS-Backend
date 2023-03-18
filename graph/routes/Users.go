package routes

import (
	"context"
	"fmt"

	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FetchUsers(jwt string) (*model.Users, error) {
	message, _ := JWTValidityCheck(jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] == "Reader" {
		panic("Unauthorized!")
	}

	// Create a temporary array of pointers for Article
	var usersStorage []model.User
	client := ConnectToMongo()
	db := client.Database("blog").Collection("Users")
	findOptions := options.Find()
	//Passing the bson.D{{}} as the filter matches documents in the collection
	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		fmt.Printf("An error has occured, could not find collection! \nFull error %s", err.Error())
	}
	defer cur.Close(context.TODO())
	var totalUsers int = 0
	// Find returns a cursor, loop through the values in the cursor
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var user model.User
		err := cur.Decode(&user)
		if err != nil {
			fmt.Printf("An error has occured, could not find decode article data! \nFull error %s", err.Error())
		}

		usersStorage = append(usersStorage, user)
		totalUsers += 1
	}
	var users = model.Users{Users: usersStorage, TotalCount: totalUsers}
	if err := cur.Err(); err != nil {
		fmt.Printf("An error has occured, could not parse cursor data! \nFull error %s", err.Error())
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	return &users, err
}
