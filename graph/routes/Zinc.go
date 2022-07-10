package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func CreateDocument(index string, data string, uuid string) {
	userName := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")

	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_doc/%s", zincBaseUrl, index, uuid)

	req, err := http.NewRequest("PUT", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

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
	fmt.Println(string(body))

}
func UpdateDocument(index string, data string, uuid string) {
	userName := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")

	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_update/%s", zincBaseUrl, index, uuid)

	req, err := http.NewRequest("POST", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

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
	fmt.Println(string(body))
}
func DeleteDocument(index string, data string, uuid string) {
	userName := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")

	zincBaseUrl := os.Getenv("ZINCBASE")
	zincDocumentUrl := fmt.Sprintf("%s/api/%s/_doc/%s", zincBaseUrl, index, uuid)

	req, err := http.NewRequest("DELETE", zincDocumentUrl, strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(userName, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

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
	fmt.Println(string(body))
}
func SearchDocuments(indexName string, searchTerm string) []byte {
	userName := os.Getenv("ZINC_FIRST_ADMIN_USER")
	password := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")
	zincBaseUrl := os.Getenv("ZINCBASE")

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

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
	fmt.Println(string(body))
}
