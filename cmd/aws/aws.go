package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"gif_wheel/wheel"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	itemParam := request.QueryStringParameters["items"]
	if itemParam == "" {
		return nil, fmt.Errorf("Missing item query parameter. Please specify the url with ?items=csv,seperated,string,of,items")
	}

	items := strings.Split(itemParam, ",")
	if len(items) == 0 || len(items) > 20 {
		return nil, fmt.Errorf("Invalid number of items.")
	}

	b := buildGif(items)
	b64 := base64.StdEncoding.EncodeToString(b)

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":   "image/gif",
			"Content-Length": fmt.Sprintf("%d", len(b64)),
		},
		Body:            b64,
		IsBase64Encoded: true,
	}, nil

}

func buildGif(items []string) []byte {
	wheel := wheel.NewWheel(60, 600, 600, 250, items)

	var b bytes.Buffer
	bb := bufio.NewWriter(&b)

	err := wheel.BuildGif(bb)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return b.Bytes()
}
