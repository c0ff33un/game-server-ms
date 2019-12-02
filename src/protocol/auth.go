package protocol

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Query struct {
	Data struct {
		User struct {
			Id     int
			Handle string
			Email  string
			Guest  bool
		}
	}
}

func TokenQuery(token string) (*Query, error) {
	url := "http://" + os.Getenv("GRAPHQL_URL") + "/graphql?query=" + url.QueryEscape(`{ user {id handle email guest} }`)
	bearer := "Bearer " + token
	req, err := http.NewRequest("GET", url, nil)
	fmt.Println("TokenQuery url:", url)
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
		return nil, err
	}
	var f Query
	err = json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		fmt.Println(r.Body)
		log.Println("Error decoding JSON.\n[ERROR] -", err)
		return nil, err
	}
	return &f, nil
}
