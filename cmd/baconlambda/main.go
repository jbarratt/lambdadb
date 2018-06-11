package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jbarratt/lambdadb/bacon"
)

var (
	b *bacon.Bacon
)

// Response is what to say back to alexa
type Response struct {
	Message string
}

// Handler is the lambda handler
func Handler() (Response, error) {
	start := b.RandomPerson()
	end := b.RandomPerson()
	path, _ := b.FindPath(start, end)

	return Response{Message: path.Prose()}, nil
}

func main() {
	b, _ = bacon.NewFromGob("bacon.gob")
	lambda.Start(Handler)
}
