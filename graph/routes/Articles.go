package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Zinc struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index     string    `json:"_index"`
			Type      string    `json:"_type"`
			ID        string    `json:"_id"`
			Score     float64   `json:"_score"`
			Timestamp time.Time `json:"@timestamp"`
			Source    struct {
				ContentData    string    `json:"ContentData"`
				DateWritten    time.Time `json:"DateWritten"`
				Description    string    `json:"Description"`
				Project        string    `json:"Project"`
				Tags           string    `json:"Tags"`
				Title          string    `json:"Title"`
				TitleCard      string    `json:"TitleCard"`
				UUID           string    `json:"UUID"`
				URL            string    `json:"Url"`
				Username       string    `json:"Username"`
				Name           string    `json:"Name"`
				ProfilePicture string    `json:"ProfilePicture"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type Total struct {
	Value int64 `json:"value"`
}

func FetchArticles(input *model.ArticlesPrivate) (*model.Articles, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if message == "Unauthorized!" || redisData["role"] == "Reader" {
		panic("Unauthorized!")
	}

	// Create a temporary array of pointers for Article
	var articlesStorage []model.Article
	client := ConnectToMongo()
	db := client.Database(fmt.Sprintf("%s", input.ProjectUUID)).Collection("articles")
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
	_, zincPassword := ZincLogin(input.ProjectUUID)
	var articles = model.Articles{Article: articlesStorage, Total: totalArticles, ZincPassword: zincPassword}
	if err := cur.Err(); err != nil {
		fmt.Printf("An error has occured, could not parse cursor data! \nFull error %s", err.Error())
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	return &articles, err
}
func FetchArticlesZinc(input *model.GetZincArticleInput) (*model.Articles, error) {
	username := input.Username
	password := input.Password
	// Create a temporary array of pointers for Article
	var articlesStorage []model.Article
	var zinc Zinc
	data := SearchDocuments(fmt.Sprintf("%s-articles", username), input.Keyword, username, password, input.PageNumber)
	zincError := json.Unmarshal(data, &zinc)
	if zincError != nil {
		panic(fmt.Errorf("error is %v", zincError))
	}

	hits := zinc.Hits.Hits
	totalArticles := 0
	var tags []model.Tag

	for hit := range hits {
		author := model.Author{Name: hits[hit].Source.Name, Profile: "", Picture: hits[hit].Source.ProfilePicture, Username: hits[hit].Source.Username}
		article := model.Article{Author: &author, ContentData: hits[hit].Source.ContentData, DateWritten: hits[hit].Source.DateWritten.String(), Description: hits[hit].Source.Description, Tags: tags, Title: hits[hit].Source.Title, TitleCard: hits[hit].Source.TitleCard, UUID: hits[hit].Source.UUID, URL: hits[hit].Source.URL}
		articlesStorage = append(articlesStorage, article)
		totalArticles += 1
	}

	var articles = model.Articles{Article: articlesStorage, Total: totalArticles}
	log.Println(articlesStorage)
	return &articles, zincError
}
func DeleteArticles(input *model.DeleteAllArticlesInput) (string, error) {
	message, _ := JWTValidityCheck(input.Jwt)
	if message == "Unauthorized!" {
		panic("Unauthorized!")
	}

	client := ConnectToMongo()
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	role := redisData["role"]
	if role == "Admin" {
		if err := client.Database(fmt.Sprintf("%s", input.ProjectUUID)).Collection("articles").Drop(context.TODO()); err != nil {
			log.Fatal(err)
		}
		zincusername, password := ZincLogin(input.ProjectUUID)
		articlesIndex := fmt.Sprintf("%s-articles", zincusername)
		imagesIndex := fmt.Sprintf("%s-images", zincusername)
		DeleteIndex(articlesIndex, zincusername, password)
		session := CreateAWSSession()
		s3sc := s3.New(session)
		iter := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
			Bucket: aws.String(imagesIndex),
		})
		DeleteArticleFolder(s3sc, iter, imagesIndex)
		var err error
		return "successfully deleted articles", err
	}
	var err error
	return "", err
}
