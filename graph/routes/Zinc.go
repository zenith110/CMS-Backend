package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateDocument(index string, data string, uuid string, userName string, password string) {
	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_doc/%s", zincBaseUrl, index, uuid)
	req, err := http.NewRequest("PUT", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured when sending data! %v", err))
	}

	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured while grabbing data! %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(fmt.Errorf("error occured while reading the data! %v", err))
	}

	log.WithFields(log.Fields{
		"document state": "Returning response",
	}).Info(fmt.Sprintf("document data: %s", string(body)))

}
func UpdateDocument(index string, data string, userName string, password string, uuid string) {
	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_update/%s", zincBaseUrl, index, uuid)

	req, err := http.NewRequest("POST", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "Returning response",
	}).Info(fmt.Sprintf("Article data: %s", string(body)))
}
func DeleteDocument(index string, data string, uuid string, userName string, password string) {
	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_doc/%s", zincBaseUrl, index, uuid)

	req, err := http.NewRequest("DELETE", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "Returning response",
	}).Info(fmt.Sprintf("Article data: %s", string(body)))
}
func SearchResults(query string, zincBaseUrl string, userName string, password string, indexName string) []byte {
	finalURL := fmt.Sprintf("%s/api/%s/_search", zincBaseUrl, indexName)

	req, err := http.NewRequest("POST", finalURL, strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}
func SearchDocuments(indexName string, searchTerm string, userName string, password string, pageNumber string) []byte {
	zincBaseUrl := os.Getenv("ZINCBASE")
	if searchTerm == "" {
		query := `{
			"search_type": "matchall",
			"_source": []
		}`
		results := SearchResults(query, zincBaseUrl, userName, password, indexName)
		return results
	}
	query := fmt.Sprintf(`{
        "search_type": "fuzzy",
        "query":
        {
            "term": "%s",
			"field": "title"
        },
        "from": %s,
        "max_results": 20,
        "_source": []
    }`, searchTerm, pageNumber)
	results := SearchResults(query, zincBaseUrl, userName, password, indexName)

	return results
}
func DeleteIndex(index string, username string, password string) {
	data := ""

	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/index/%s", zincBaseUrl, index)

	req, err := http.NewRequest("DELETE", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "Returning response",
	}).Info(fmt.Sprintf("Article data: %s", string(body)))
}

func ZincLogin(uuid string) (string, string) {
	username := uuid
	client := ConnectToMongo()
	collection := client.Database("zinc").Collection("users")
	var zincUser model.ZincUser

	//Passing the bson.D{{}} as the filter matches documents in the collection
	zincErr := collection.FindOne(context.TODO(), bson.M{"username": uuid}).Decode(&zincUser)
	if zincErr != nil {
		log.Fatal(zincErr)
	}

	password, _ := Decrypt(zincUser.Password)
	defer CloseClientDB()
	return username, password
}

func CreateZincUser(username string, password string, email string) {
	zincUsername := os.Getenv("ZINC_FIRST_ADMIN_USER")
	zincPassword := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")
	zincBaseUrl := os.Getenv("ZINCBASE")
	zincData := fmt.Sprintf(`{
			"_id": "%s",
			"name": "%s",
			"role": "Admin",
			"password": "%s"
		}`, username, email, password)
	zincDocumentUrl := fmt.Sprintf("%s/api/user", zincBaseUrl)
	req, err := http.NewRequest("POST", zincDocumentUrl, strings.NewReader(zincData))
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured when sending data! %v", err))
	}

	req.SetBasicAuth(zincUsername, zincPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured while grabbing data! %v", err))
	}
	defer resp.Body.Close()

	_, userCreationErr := io.ReadAll(resp.Body)
	if userCreationErr != nil {
		log.Fatal(fmt.Errorf("error occured while reading the data! %v", err))
	}
}
func DeleteZincUser(uuid string, zincUsername string, zincPassword string) {
	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/user/%s", zincBaseUrl, uuid)
	req, err := http.NewRequest("DELETE", zincDocumentUrl, strings.NewReader(""))
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured when sending data! %v", err))
	}

	req.SetBasicAuth(zincUsername, zincPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("error has occured while grabbing data! %v", err))
	}
	defer resp.Body.Close()
}
