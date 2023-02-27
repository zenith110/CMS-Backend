package routes

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func UploadImageGallery(file *model.File) (string, error) {
// 	var err error
// 	return UploadToGallery(file), err
// }

func UploadImageDB(image model.Image, url string, username string, project string) (model.Image, error) {
	client := ConnectToMongo()
	collection := client.Database(username).Collection("images")
	res, err := collection.InsertOne(context.TODO(), image)
	if err != nil {
		log.Fatal(err)
	}
	defer log.WithFields(log.Fields{
		"article state": "finished insertion",
	}).Info(fmt.Sprintf("Inserted a single document: %s", res.InsertedID))
	defer CloseClientDB()
	return image, err
}
func GalleryFindImages(jwt string, username string) (*model.GalleryImages, error) {
	message, _ := JWTValidityCheck(jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}
	var err error
	// Create a temporary array of pointers for Article
	var imagesStorage []model.Image
	client := ConnectToMongo()
	db := client.Database(username).Collection("images")
	findOptions := options.Find()
	//Passing the bson.D{{}} as the filter matches documents in the collection
	cur, err := db.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		fmt.Printf("An error has occured, could not find collection! \nFull error %s", err.Error())
	}
	defer cur.Close(context.TODO())
	var totalImages int = 0
	// Find returns a cursor, loop through the values in the cursor
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var image model.Image
		err := cur.Decode(&image)
		if err != nil {
			fmt.Printf("An error has occured, could not find decode article data! \nFull error %s", err.Error())
		}

		imagesStorage = append(imagesStorage, image)
		totalImages += 1
	}
	var images = model.GalleryImages{Images: imagesStorage, Total: totalImages}
	if err := cur.Err(); err != nil {
		fmt.Printf("An error has occured, could not parse cursor data! \nFull error %s", err.Error())
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	defer CloseClientDB()
	return &images, err
}
