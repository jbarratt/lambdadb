package main

import (
	"fmt"
	"time"

	alexa "github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jbarratt/lambdadb/bacon"
)

var (
	b *bacon.Bacon
)

// DispatchIntents dispatches each intent to the right handler
func DispatchIntents(request alexa.Request) alexa.Response {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "GetBaconPath":
		response = handleBacon(request)
	case alexa.HelpIntent:
		response = handleHelp()
	}

	return response
}

// Handler is the lambda hander
func Handler(request alexa.Request) (alexa.Response, error) {
	return DispatchIntents(request), nil
}

// Handler is the lambda handler
func handleBacon(r alexa.Request) alexa.Response {
	if len(r.Body.Intent.Slots["fromActor"].Value) < 3 {
		return alexa.NewSimpleResponse("Missing From", "Hm, I only heard one actor name. Try something like 'ask Bacon Guru to link James Dean to John Malkovitch'")
	}
	loadStart := time.Now()

	fmt.Printf("fromActor: %s\n", r.Body.Intent.Slots["fromActor"].Value)
	fmt.Printf("toActor: %s\n", r.Body.Intent.Slots["toActor"].Value)

	findFrom := time.Now()
	start := b.FindPerson(r.Body.Intent.Slots["fromActor"].Value)
	fmt.Printf("Time to find 'From': %s\n", time.Since(findFrom))

	findTo := time.Now()
	end := b.FindPerson(r.Body.Intent.Slots["toActor"].Value)
	fmt.Printf("Time to find 'To': %s\n", time.Since(findTo))

	findPath := time.Now()
	path, err := b.FindPath(start, end)
	fmt.Printf("Time to find path: %s\n", time.Since(findPath))

	fmt.Printf("Skill Duration %s\n", time.Since(loadStart))

	if err == nil {
		return alexa.NewSimpleResponse("Path Found", path.Prose())
	}
	return alexa.NewSimpleResponse("No Path Found", "Unable to connect those actors.")
}

func handleHelp() alexa.Response {
	return alexa.NewSimpleResponse("Help for Bacon Guru", "You can say things like ask Bacon Guru to link Tom Cruise to Tom Hardy")
}

func main() {
	b, _ = bacon.NewFromGob("bacon.gob")
	lambda.Start(Handler)
}
