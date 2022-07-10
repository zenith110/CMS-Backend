package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zenith110/portfilo/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Zinc struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Hits     Hits  `json:"hits"`
}

type Hits struct {
	Total    Total `json:"total"`
	MaxScore int64 `json:"max_score"`
	Hits     []Hit `json:"hits"`
}

type Hit struct {
	Index     string `json:"_index"`
	Type      string `json:"_type"`
	ID        string `json:"_id"`
	Score     int64  `json:"_score"`
	Timestamp string `json:"@timestamp"`
	Source    Source `json:"_source"`
}

type Source struct {
	Author      string `json:"Author"`
	ContentData string `json:"ContentData"`
	DateWritten string `json:"DateWritten"`
	Description string `json:"Description"`
	Tags        string `json:"Tags"`
	Title       string `json:"Title"`
	TitleCard   string `json:"TitleCard"`
	UUID        string `json:"UUID"`
	URL         string `json:"Url"`
}

type Total struct {
	Value int64 `json:"value"`
}

func FetchArticles() (*model.Articles, error) {
	var err error
	// Create a temporary array of pointers for Article
	var articlesStorage []model.Article
	client := ConnectToMongo()
	db := client.Database("blog").Collection("articles")
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
		var article model.Article
		err := cur.Decode(&article)
		if err != nil {
			fmt.Printf("An error has occured, could not find decode article data! \nFull error %s", err.Error())
		}

		articlesStorage = append(articlesStorage, article)
		totalArticles += 1
	}
	var articles = model.Articles{Article: articlesStorage, Total: totalArticles}
	if err := cur.Err(); err != nil {
		fmt.Printf("An error has occured, could not parse cursor data! \nFull error %s", err.Error())
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	return &articles, err
}
func FetchArticlesZinc(keyword string) (*model.Articles, error) {
	var err error
	// Create a temporary array of pointers for Article
	var articlesStorage []model.Article
	var zinc Zinc
	data := SearchDocuments("articles", keyword)
	zincError := json.Unmarshal(data, &zinc)
	if zincError != nil {
		fmt.Errorf("error is %v", zincError)
	}

	hits := zinc.Hits.Hits
	totalArticles := 0
	var tags []model.Tag

	for hit := range hits {
		author := model.Author{Name: hits[hit].Source.Author, Profile: "", Picture: ""}
		article := model.Article{Author: &author, ContentData: hits[hit].Source.ContentData, DateWritten: hits[hit].Source.DateWritten, Description: hits[hit].Source.Description, Tags: tags, Title: hits[hit].Source.Title, TitleCard: hits[hit].Source.TitleCard, UUID: hits[hit].Source.UUID, URL: hits[hit].Source.URL}
		articlesStorage = append(articlesStorage, article)
		totalArticles += 1
	}

	// var articles model.Articles
	var articles = model.Articles{Article: articlesStorage, Total: totalArticles}
	return &articles, err
}
func DeleteArticles() (*model.Article, error) {
	client := ConnectToMongo()
	fmt.Print("Connected to mongodb!")
	if err := client.Database("blog").Collection("articles").Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	DeleteIndex("articles")
	var article model.Article
	var err error
	return &article, err
}
