package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/jbarratt/lambdadb/bacon"
	"github.com/urfave/cli"
	pb "gopkg.in/cheggaaa/pb.v2"
)

func main() {

	rand.Seed(time.Now().UnixNano())
	// defer profile.Start(profile.MemProfile).Stop()
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
		{
			Name:  "monte",
			Usage: "Find common linkers",
			Action: func(c *cli.Context) error {
				monteCarlo()
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

	fmt.Printf("Number of nodes: %d\n", bacon.Graph.Order())
	fmt.Printf("%d films and %d people\n", bacon.Graph.Order()-len(bacon.People), len(bacon.People))

}

func monteCarlo() {
	b, err := bacon.NewFromGob("data/bacon.gob")
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}
	counts := make([]uint, b.Graph.Order(), b.Graph.Order())

	iterations := 400000
	// iterations := 100000
	bar := pb.StartNew(iterations)
	for i := 0; i < iterations; i++ {
		bar.Increment()
		startNode := b.RandomPerson()
		endNode := b.RandomPerson()
		path, err := b.FindPath(startNode, endNode)
		if err != nil {
			continue
		}
		for i := 0; i < len(path); i++ {
			counts[path[i].Node] += 1
		}
	}
	bar.Finish()
	// sort this data
	type score struct {
		name  string
		count uint
	}
	people := make([]score, 100000)
	for id, count := range counts {
		if count > 0 && b.NodeInfo[id].IsPerson {
			people = append(people, score{name: b.NodeInfo[id].Name, count: count})
		}
	}
	sort.Slice(people, func(i, j int) bool {
		return people[i].count > people[j].count
	})
	f, _ := os.Create("people.csv")
	for _, person := range people {
		fmt.Fprintf(f, "%s,%d\n", person.name, person.count)
	}
	f.Close()

	movies := make([]score, 100000)
	for id, count := range counts {
		if count > 0 && !b.NodeInfo[id].IsPerson {
			movies = append(movies, score{name: b.NodeInfo[id].Name, count: count})
		}
	}
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].count > movies[j].count
	})
	f, _ = os.Create("movies.csv")
	for _, movie := range movies {
		fmt.Fprintf(f, "%s,%d\n", movie.name, movie.count)
	}
	f.Close()
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
