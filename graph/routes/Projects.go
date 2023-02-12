package routes

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateProject(jwt string, email string, password string, role string, input *model.CreateProjectInput) (*model.Project, error) {
	message, _ := JWTValidityCheck(jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	user := AuthenticateNonReaders(email, password, jwt, role)
	if user.Role != "Creator" || user.Role == "Admin" {
		client := ConnectToMongo()
		collection := client.Database(email).Collection("projects")
		var articles model.Articles
		project := model.Project{Name: input.Name, UUID: input.UUID, Articles: &articles, Authur: input.Author}
		_, err := collection.InsertOne(context.TODO(), project)
		if err != nil {
			var emptyProject model.Project
			return &emptyProject, err
		}
		return &project, nil
	}
	var emptyProject model.Project
	defer CloseClientDB()
	return &emptyProject, nil
}

func GetProjects(jwt string, email string, password string) (*model.Projects, error) {
	message, _ := JWTValidityCheck(jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	var err error
	// Create a temporary array of pointers for Article
	var projectsStorage []model.Project
	client := ConnectToMongo()
	db := client.Database(email).Collection("projects")
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
func DeleteProject(jwt string, email string, password string, project string, uuid string) (string, error) {
	message, _ := JWTValidityCheck(jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	client := ConnectToMongo()
	collection := client.Database(email).Collection("projects")
	deleteResult, deleteError := collection.DeleteOne(context.TODO(), bson.M{"uuid": uuid})
	if deleteResult.DeletedCount == 0 {
		log.Fatal("Error on deleting data ", deleteError)
	}
	defer CloseClientDB()
	return fmt.Sprintf("Deleted project %s", project), deleteError
}
