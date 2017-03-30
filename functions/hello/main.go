package main

import (
	"context"
	"encoding/json"
	"os"
	"reflect"

	elastic "gopkg.in/olivere/elastic.v5"

	apex "github.com/apex/go-apex"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/bluele/slack"
	"github.com/deoxxa/aws_signing_client"
)

type message struct {
	Hello string `json:"hello"`
}

type Tweet struct {
	User string `json:"user"`
	Text string `json:"text"`
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		var m message

		token := os.Getenv("SLACK_TOKEN")
		channelName := os.Getenv("SLACK_CHANNEL_NAME")

		creds := credentials.NewStaticCredentials(os.Getenv("AWS_ES_ACCESS_KEY_ID"), os.Getenv("AWS_ES_SECRET_ACCESS_KEY"), "")
		signer := v4.NewSigner(creds)
		awsClient, _ := aws_signing_client.New(signer, nil, "es", "ap-northeast-1")

		client, _ := elastic.NewClient(
			elastic.SetURL(os.Getenv("AWS_ES_ENDPOINT")),
			elastic.SetScheme("https"),
			elastic.SetHttpClient(awsClient),
			elastic.SetSniff(false),
		)

		api := slack.New(token)

		query := elastic.NewTermQuery("text", "mackerel")
		searchResult, _ := client.Search().Index("twitter_public_timeline").Query(query).Do(context.Background())

		var ttyp Tweet
		for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
			if t, ok := item.(Tweet); ok {
				err := api.ChatPostMessage(channelName, t.User+": "+t.Text, nil)
				if err != nil {
					panic(err)
				}
			}
		}

		return m, nil
	})
}
