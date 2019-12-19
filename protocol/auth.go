package protocol

import (
	"context"
	"log"
	"os"

	"github.com/machinebox/graphql"
)

type Response struct {
	Me struct {
		Id     string
		Handle string
		Email  string
	}
}

func GetUser(token string) (Response, error) {
	url := os.Getenv("GRAPHQL_URL")
	if url == "" {
		panic("please provide a graphql url")
	}
	log.Printf("Graphql url: %v", url)
	client := graphql.NewClient(url)
	req := graphql.NewRequest(`
		query {
			me { 
				id
				handle
				email
			}
		}
	`)
	log.Printf("token: %v", token)
	req.Header.Set("Authorization", "Bearer "+token)
	//req.Header.Set("Accept", "application/json")
	ctx := context.TODO()
	respData := Response{}
	err := client.Run(ctx, req, &respData)
	return respData, err
}
