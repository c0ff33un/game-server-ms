package protocol

import (
	"context"
	"os"

	"github.com/machinebox/graphql"
)

type Response struct {
	User struct {
		Id     int
		Handle string
		Email  string
	}
}

func GetUser(token string) (Response, error) {
	client := graphql.NewClient(os.Getenv("GRAPHQL_URL"))
	req := graphql.NewRequest(`
		query {
			user { 
				id
				handle
				email
			}
		}
	`)
	req.Header.Set("Authorization", "Bearer "+token)
	//req.Header.Set("Accept", "application/json")
	ctx := context.TODO()
	respData := Response{}
	err := client.Run(ctx, req, &respData)
	return respData, err
}
