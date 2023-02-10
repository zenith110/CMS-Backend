package routes

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateArticle(input *model.CreateArticleInfo) (*model.Article, error) {
	/*
		Creates a temporary array for the article model, loop through the contents of the input for all the tag data
	*/
	if input.Jwt == "" {
		log.Panic("JWT is invalid! It's empty!")
		panic("JWT is empty!")
	}
	var tags []model.Tag
	var tagsString []string
	for tagData := 0; tagData < len(input.Tags); tagData++ {
		tag := model.Tag{
			Language: *input.Tags[tagData].Name,
		}
		tags = append(tags, tag)
		tagsString = append(tagsString, *input.Tags[tagData].Name)
	}
	imageURL := UploadFileToS3(input)
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("articles")
	author := model.Author{Name: *input.Author, Profile: "", Picture: ""}
	article := model.Article{Title: *input.Title, Author: &author, ContentData: *input.ContentData, DateWritten: *input.DateWritten, URL: *input.URL, Description: *input.Description, UUID: *input.UUID, Tags: tags, TitleCard: imageURL}
	res, err := collection.InsertOne(context.TODO(), article)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "pre-insert mongo data",
	}).Info("Article has been created, inserting into zinc!")

	zincData := fmt.Sprintf(`{
		"Title":       "%s",
		"Author":      "%s",
		"ContentData": "%s",
		"DateWritten": "%s",
		"Url":         "%s",
		"Description": "%s",
		"UUID":        "%s",
		"TitleCard":   "%s",
		"Tags":        "%s"
	}`, *input.Title, *input.Author, *input.ContentData, *input.DateWritten, *input.URL, *input.Description, *input.UUID, imageURL, strings.Join(tagsString, ","))

	log.WithFields(log.Fields{
		"article state": "created mongodb instance",
	}).Info("Article has been created, inserting into zinc!")

	CreateDocument("articles", zincData, *input.UUID)
	log.WithFields(log.Fields{
		"article state": "finished insertion",
	}).Info(fmt.Sprintf("Inserted a single document: %s", res.InsertedID))
	return &article, err
}
func DeleteArticle(bucket *model.DeleteBucketInfo) (*model.Article, error) {
	if bucket.Jwt == "" {
		log.Panic("JWT is invalid!")
		panic("JWT is empty!")
	}
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("articles")
	article := model.Article{UUID: *bucket.UUID}
	DeleteArticleBucket(*bucket.BucketName)
	deleteResult, deleteError := collection.DeleteOne(context.TODO(), bson.M{"uuid": *bucket.UUID})
	if deleteResult.DeletedCount == 0 {
		log.Fatal("Error on deleting data ", deleteError)
	}
	zincData := fmt.Sprintf(`{
		"UUID":        "%s"
	}`, *bucket.UUID)
	DeleteDocument("articles", zincData, *bucket.UUID)
	return &article, deleteError
}
func FindArticle(title *string, jwt *string) (*model.Article, error) {
	if jwt != nil {
		log.Panic("JWT is invalid!")
		panic("JWT is invalid!")
	}
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("articles")
	var article model.Article

	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"url": title}).Decode(&article)
	if err != nil {
		log.Fatal(err)
	}
	return &article, err
}
func UpdateArticle(input *model.UpdatedArticleInfo) (*model.Article, error) {
	if input.Jwt == "" {
		log.Panic("JWT is empty!")
		panic("JWT is empty!")
	}
	var tags []model.Tag
	var tagsString []string
	for tagData := 0; tagData < len(input.Tags); tagData++ {
		tag := model.Tag{
			Language: *input.Tags[tagData].Name,
		}
		tags = append(tags, tag)
		tagsString = append(tagsString, *input.Tags[tagData].Name)
	}
	imageURL := UploadUpdatedFileToS3(input)
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("articles")

	filter := bson.M{"uuid": input.UUID}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "title", Value: *input.Title}, primitive.E{Key: "TitleCard", Value: imageURL}, primitive.E{Key: "contentData", Value: *input.ContentData}, primitive.E{Key: "URL", Value: *input.URL}, primitive.E{Key: "Description", Value: input.Description}, primitive.E{Key: "tags", Value: tags},
	}}}
	var article model.Article
	_, err := collection.UpdateOne(
		context.TODO(),
		filter,
		update,
	)
	if err != nil {
		panic(fmt.Errorf("error has occured: %v", err))
	}

	zincData := fmt.Sprintf(`{
		"Title":       "%s",
		"Author":      "%s",
		"ContentData": "%s",
		"DateWritten": "%s",
		"Url":         "%s",
		"Description": "%s",
		"UUID":        "%s",
		"TitleCard":   "%s",
		"Tags":        "%s"
	}`, *input.Title, *input.Author, *input.ContentData, *input.DateWritten, *input.URL, *input.Description, *input.UUID, imageURL, strings.Join(tagsString, ","))
	UpdateDocument("articles", zincData, *input.UUID)
	return &article, err
}
