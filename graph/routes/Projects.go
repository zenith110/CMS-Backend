package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
Creates a AWS bucket given project paramaters
*/
func CreateProject(input *model.CreateProjectInput) (*model.Project, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] == "Reader" {
		panic("Unauthorized!")
	}
	client := ConnectToMongo()
	collection := client.Database(redisData["username"]).Collection("projects")
	var articles model.Articles
	project := model.Project{Name: input.Name, UUID: input.UUID, Articles: &articles, Author: redisData["username"], Description: input.Description, EncryptionKey: os.Getenv("ENCRYPTIONKEY")}
	_, err := collection.InsertOne(context.TODO(), project)
	if err != nil {
		var emptyProject model.Project
		return &emptyProject, err
	}
	session := CreateAWSSession()
	s3sc := s3.New(session)
	bucketName := fmt.Sprintf("%s-images", input.UUID)
	bucketExist := CheckIfBucketExist(s3sc, bucketName)
	password := randstr.String(36)
	encryptedPassword, err := Encrypt(password)
	if err != nil {
		panic(fmt.Errorf("\nError is %v", err))
	}
	if bucketExist == true {
		return &project, nil
	}
	CreateProjectBucket(s3sc, bucketName)
	CreateZincUser(input.UUID, password, "")
	zinccollection := client.Database("zinc").Collection("users")
	zincuser := model.ZincUser{Username: input.UUID, Password: encryptedPassword}
	_, zincerr := zinccollection.InsertOne(context.TODO(), zincuser)
	if zincerr != nil {
		var emptyProject model.Project
		return &emptyProject, err
	}

	return &project, err
}

func GetProjects(input *model.GetProjectType) (*model.Projects, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] == "Reader" {
		panic("Unauthorized!")
	}
	var err error

	// Create a temporary array of pointers for projects
	var projectsStorage []model.Project
	client := ConnectToMongo()
	db := client.Database(redisData["username"]).Collection("projects")
	findOptions := options.Find()
	//Passing the bson.D{{}} as the filter matches documents in the collection
	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		fmt.Printf("An error has occured, could not find collection! \nFull error %s", err.Error())
	}
	defer cur.Close(context.TODO())
	var totalArticles int = 0
	// Find returns a cursor, loop through the values in the cursor
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var project model.Project
		err := cur.Decode(&project)
		if err != nil {
			fmt.Printf("An error has occured, could not find decode article data! \nFull error %s", err.Error())
		}

		projectsStorage = append(projectsStorage, project)
		totalArticles += 1
	}
	var projects = model.Projects{Projects: projectsStorage}
	if err := cur.Err(); err != nil {
		fmt.Printf("An error has occured, could not parse cursor data! \nFull error %s", err.Error())
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	defer CloseClientDB()
	return &projects, err
}
func DeleteProject(input *model.DeleteProjectType) (string, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] != "Admin" {
		panic("User cannot perform this action!")
	}
	username, password := ZincLogin(input.UUID)
	client := ConnectToMongo()
	collection := client.Database(redisData["username"]).Collection("projects")
	deleteResult, deleteError := collection.DeleteOne(context.TODO(), bson.M{"uuid": input.UUID})
	if deleteResult.DeletedCount == 0 {
		log.Fatal("Error on deleting data ", deleteError)
	}
	zinccollection := client.Database("zinc").Collection("users")
	deleteZincResult, deleteZincError := zinccollection.DeleteOne(context.TODO(), bson.M{"username": input.UUID})
	if deleteZincResult.DeletedCount == 0 {
		log.Fatal("Error on deleting data ", deleteZincError)
	}
	defer CloseClientDB()
	bucketName := fmt.Sprintf("%s-images", username)
	session := CreateAWSSession()
	// Makes an s3 service client
	s3sc := s3.New(session)
	DeleteBucket(s3sc, bucketName)
	DeleteIndex(fmt.Sprintf("%s-articles", username), username, password)
	DeleteIndex(bucketName, username, password)

	// Drop the image collection for the project
	if err := client.Database(username).Collection("images").Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	if err := client.Database(username).Collection("articles").Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	DeleteZincUser(input.UUID, username, password)
	var err error
	return fmt.Sprintf("Deleted project %s", input.Project), err
}

func DeleteProjects(input *model.DeleteAllProjects) (string, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	username := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	client := ConnectToMongo()
	if err := client.Database(redisData["username"]).Collection("projects").Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	DeleteAllProjectsBuckets(redisData["username"])
	DeleteIndex("*-articles", username, password)
	var err error
	return "", err
}
