package routes

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateProject(input *model.CreateProjectInput) (*model.Project, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}

	client := ConnectToMongo()
	collection := client.Database(input.Username).Collection("projects")
	var articles model.Articles
	project := model.Project{Name: input.Name, UUID: input.UUID, Articles: &articles, Author: input.Author, Description: input.Description}
	_, err := collection.InsertOne(context.TODO(), project)
	if err != nil {
		var emptyProject model.Project
		return &emptyProject, err
	}

	return &project, nil

}

func GetProjects(input *model.GetProjectType) (*model.Projects, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	var err error
	// Create a temporary array of pointers for Article
	var projectsStorage []model.Project
	client := ConnectToMongo()
	db := client.Database(input.Username).Collection("projects")
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
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	client := ConnectToMongo()
	collection := client.Database(input.Username).Collection("projects")
	deleteResult, deleteError := collection.DeleteOne(context.TODO(), bson.M{"uuid": input.UUID})
	if deleteResult.DeletedCount == 0 {
		log.Fatal("Error on deleting data ", deleteError)
	}
	defer CloseClientDB()
	bucketName := fmt.Sprintf("%s-%s-images", input.Username, input.UUID)
	DeleteBucket(bucketName)
	return fmt.Sprintf("Deleted project %s", input.Project), deleteError
}
func DeleteProjects(input *model.DeleteAllProjects) (string, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	client := ConnectToMongo()
	if err := client.Database(input.Username).Collection("projects").Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	DeleteAllProjectsBuckets(input.Username)
	var err error
	return "", err
}
