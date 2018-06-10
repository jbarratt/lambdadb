package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jbarratt/lambdadb/bacon"
	"github.com/urfave/cli"
)

func main() {

	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "path",
			Usage: "Find a path between 2 people",
			Action: func(c *cli.Context) error {
				findPath(c.Args().Get(0), c.Args().Get(1))
				return nil
			},
		},
		{
			Name:  "serialize",
			Usage: "convert the json db to higher perf",
			Action: func(c *cli.Context) error {
				reSerialize()
				return nil
			},
		},
		{
			Name:  "random",
			Usage: "Find a random path",
			Action: func(c *cli.Context) error {
				randomPath()
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Load the data in JSON and dump it back out in something sexier
func reSerialize() {
	loadStart := time.Now()
	bacon, err := bacon.NewFromJSON("data/bacon.json")
	log.Printf("Loading data took %s", time.Since(loadStart))
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}

	encodeStart := time.Now()
	file, err := os.Create("data/bacon.gob")

	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(bacon)
	}

	file.Close()
	log.Printf("Gobbing data took %s", time.Since(encodeStart))
}

func randomPath() {
	b, err := bacon.NewFromGob("data/bacon.gob")
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}

	for i := 0; i < 12; i++ {
		startNode := b.RandomPerson()
		endNode := b.RandomPerson()
		path, err := b.FindPath(startNode, endNode)

		if err != nil {
			fmt.Printf("No path found\n")
		} else {
			fmt.Printf("%s\n\n", path.Prose())
		}
	}
}

func findPath(start string, end string) {
	loadStart := time.Now()
	b, err := bacon.NewFromGob("data/bacon.gob")
	log.Printf("Loading data took %s", time.Since(loadStart))
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}
	startNode := b.FindPerson(start)
	endNode := b.FindPerson(end)

	fmt.Printf("Start Node: %v\n", b.NodeInfo[startNode])
	fmt.Printf("End Node: %v\n", b.NodeInfo[endNode])

	searchStart := time.Now()
	path, err := b.FindPath(startNode, endNode)
	log.Printf("Search took %s", time.Since(searchStart))
	if err != nil {
		fmt.Printf("No path found\n")
	} else {
		fmt.Println(path.Prose())
	}
}
