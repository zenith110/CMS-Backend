package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
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
	filter := bson.M{"username": bson.M{"$ne": os.Getenv("ADMINUSER")}}
	findOptions := options.Find()
	//Passing the bson.D{{}} as the filter matches documents in the collection
	cur, err := db.Find(context.TODO(), filter, findOptions)
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

func DeleteUser(input *model.DeleteUser) (string, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] == "Reader" {
		panic("Unauthorized!")
	}
	username := redisData["username"]
	role := redisData["role"]
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("Users")
	var user model.User

	//Passing the bson.D{{}} as the filter matches documents in the collection
	userErr := collection.FindOne(context.TODO(), bson.M{"uuid": input.UUID}).Decode(&user)
	if userErr != nil {
		Info(fmt.Sprint("User could not be found!"))
	}
	if role == "Admin" {
		session := CreateAWSSession()
		s3sc := s3.New(session)
		bucketName := "graphql-cms-profilepics"
		if user.Username != os.Getenv("ADMINUSER") {
			iter := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
				Bucket: aws.String(bucketName),
				Prefix: &username,
			})

			DeleteArticleFolder(s3sc, iter, bucketName)
			deleteResult, deleteError := collection.DeleteOne(context.TODO(), bson.M{"uuid": input.UUID})
			if deleteResult.DeletedCount == 0 {
				log.Fatal("Error on deleting data ", deleteError)
			}

			return fmt.Sprintf("Successful in deleting %s", user.Username), deleteError
		}
		var err error
		return fmt.Sprintf("Not possible to delete %s!", user.Username), err
	}
	var err error
	return "", err
}

func DeleteAllUsers(jwt string) (string, error) {
	message, _ := JWTValidityCheck(jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}

	client := ConnectToMongo()
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(jwt, redisClient)
	username := redisData["username"]
	if username == os.Getenv("ADMINUSER") {
		filter := bson.M{"username": bson.M{"$ne": os.Getenv("ADMINUSER")}}
		deleteResults, deleteError := client.Database("blog").Collection("Users").DeleteMany(context.TODO(), filter)
		if deleteResults.DeletedCount == 0 {
			log.Fatal("Error on deleting data ", deleteError)
		}
		session := CreateAWSSession()
		s3sc := s3.New(session)
		imagesIndex := "graphql-cms-profilepics"
		iter := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
			Bucket: aws.String(imagesIndex),
		})
		DeleteArticleFolder(s3sc, iter, imagesIndex)
		var err error
		return "successfully deleted all users but base user!", err
	}
	var err error
	return "Could not delete all users.", err
}
