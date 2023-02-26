package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
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
func UpdateDocument(index string, data string, uuid string, userName string, password string) {

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
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "Returning response",
	}).Info(fmt.Sprintf("Article data: %s", string(body)))
}
func SearchDocuments(indexName string, searchTerm string, userName string, password string) []byte {
	zincBaseUrl := os.Getenv("ZINCBASE")
	if searchTerm == "" {

	}
	query := fmt.Sprintf(`{
        "search_type": "match",
        "query":
        {
            "term": "%s"
        },
        "from": 0,
        "max_results": 20,
        "_source": []
    }`, searchTerm)
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
func DeleteIndex(index string) {
	data := ""
	userName := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")

	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/index/%s", zincBaseUrl, index)

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
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"article state": "Returning response",
	}).Info(fmt.Sprintf("Article data: %s", string(body)))
}
