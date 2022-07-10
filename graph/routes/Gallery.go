package routes

import (
	"context"
	"fmt"
	"log"

	"github.com/zenith110/portfilo/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UploadImageGallery(file *model.File) (string, error) {
	var err error
	return UploadToGallery(file), err
}

func UploadImageDB(image model.Image, url string) (model.Image, error) {
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("images")
	res, err := collection.InsertOne(context.TODO(), image)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", res.InsertedID)
	return image, err
}
func GalleryFindImages() (*model.GalleryImages, error) {
	var err error
	// Create a temporary array of pointers for Article
	var imagesStorage []model.Image
	client := ConnectToMongo()
	db := client.Database("blog").Collection("images")
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

	return &images, err
}
