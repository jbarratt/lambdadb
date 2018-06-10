package main

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/arbovm/levenshtein"
	"github.com/shawnsmithdev/zermelo/zuint32"
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
	bacon, err := LoadBaconJSON("data/bacon.json")
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
	bacon, err := LoadBaconGob("data/bacon.gob")
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}

	for i := 0; i < 10; i++ {
		startNode := bacon.RandomPerson()
		endNode := bacon.RandomPerson()
		path, err := bacon.BreadthFirstSearch(startNode, endNode)

		if err != nil {
			fmt.Printf("No path found\n")
		} else {
			fmt.Println(path.Prose())
		}
	}
}

func findPath(start string, end string) {
	loadStart := time.Now()
	bacon, err := LoadBaconGob("data/bacon.gob")
	log.Printf("Loading data took %s", time.Since(loadStart))
	if err != nil {
		log.Fatalf("FAIL: %v\n", err)
	}
	startNode := bacon.find(start)
	endNode := bacon.find(end)

	fmt.Printf("Start Node: %v\n", bacon.NodeInfo[startNode])
	fmt.Printf("End Node: %v\n", bacon.NodeInfo[endNode])

	searchStart := time.Now()
	path, err := bacon.BreadthFirstSearch(startNode, endNode)
	log.Printf("Search took %s", time.Since(searchStart))
	if err != nil {
		fmt.Printf("No path found\n")
	} else {
		fmt.Println(path.Prose())
	}
}

// Node represents a graph node
type Node = uint32

// Graph is the graph storage data
type Graph struct {
	// List is the list of neighbors
	List []Node   `json:"list"`
	Span []uint64 `json:"span"`
}

// NodeInfo contains extra info about a node
type NodeInfo struct {
	Name   string
	Kind   string `json:"type"`
	TmdbID uint32 `json:"tmdb_id"`
}

// People enables lookup of people by name to ID
type People map[string]Node

// Bacon stores needed data for computing bacon numbers
type Bacon struct {
	Graph    Graph      `json:"graph"`
	NodeInfo []NodeInfo `json:"node_data"`
	People   People     `json:"people"`
}

type Path []NodeInfo

// LoadBaconJSON is like NewBacon with a JSON path
func LoadBaconJSON(path string) (*Bacon, error) {
	bacon := new(Bacon)
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &bacon)
	if err != nil {
		return nil, err
	}
	return bacon, nil
}

// LoadBaconGob is like NewBacon with a JSON path
func LoadBaconGob(path string) (*Bacon, error) {
	bacon := new(Bacon)
	gobFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(gobFile)
	err = decoder.Decode(bacon)
	if err != nil {
		return nil, err
	}

	gobFile.Close()
	return bacon, nil
}

func (b *Bacon) find(name string) Node {
	node, ok := b.People[name]
	if ok {
		return node
	}
	fmt.Println("Name not found, using most similar")
	minDist := math.MaxInt32
	minNode := Node(0)
	for person, node := range b.People {
		dist := levenshtein.Distance(person, name)
		if dist < minDist {
			minDist = dist
			minNode = node
		}
	}
	return minNode
}

// Neighbors returns a slice of the neighborhs of a graph node
func (graph *Graph) Neighbors(n Node) []Node {
	start, end := graph.Span[n], graph.Span[n+1]
	return graph.List[start:end]
}

// Order returns the order of the graph
func (graph *Graph) Order() int {
	return len(graph.List)
}

// from package 'search'
// https://raw.githubusercontent.com/egonelbre/a-tale-of-bfs/master/09_unroll_8_4/nodeset.go

const (
	bucketBits = 5
	bucketSize = 1 << 5
	bucketMask = bucketSize - 1
)

// NodeSet is a bit vector tracking set of nodes
type NodeSet []uint32

// NewNodeSet constructs a new bit vector node tracker
func NewNodeSet(size int) NodeSet {
	return NodeSet(make([]uint32, (size+31)/32))
}

// Offset figures out the right bucket/bit offset for a node
func (set NodeSet) Offset(node Node) (bucket, bit uint32) {
	bucket = uint32(node >> bucketBits)
	bit = uint32(1 << (node & bucketMask))
	return bucket, bit
}

// Add sets the right bucket and bit
func (set NodeSet) Add(node Node) {
	bucket, bit := set.Offset(node)
	set[bucket] |= bit
}

// Contains checks the bucket/bit for a bitset to see if it's set
func (set NodeSet) Contains(node Node) bool {
	bucket, bit := set.Offset(node)
	return set[bucket]&bit != 0
}

// BreadthFirstSearch returns bacon number of source to dest
// Returns a slice of nodes representing the path from source to dest
func (b *Bacon) BreadthFirstSearch(source Node, dest Node) (Path, error) {

	g := b.Graph

	visited := NewNodeSet(g.Order())

	currentLevel := make([]Node, 0, g.Order())
	nextLevel := make([]Node, 0, g.Order())
	parentNode := make([]Node, g.Order(), g.Order())

	visited.Add(source)
	currentLevel = append(currentLevel, source)

	for len(currentLevel) > 0 {
		for _, node := range currentLevel {
			for _, neighbor := range g.Neighbors(node) {
				if !visited.Contains(neighbor) {
					nextLevel = append(nextLevel, neighbor)
					visited.Add(neighbor)
					parentNode[neighbor] = node
				}
				if neighbor == dest {
					return NewPath(source, dest, parentNode, b), nil
				}
			}
		}

		// Make it more likely that each page of the node list will be
		// loaded into cache only once
		zuint32.SortBYOB(nextLevel, currentLevel[:cap(currentLevel)])

		currentLevel = currentLevel[:0:cap(currentLevel)]
		currentLevel, nextLevel = nextLevel, currentLevel
	}
	return nil, errors.New("No path found")
}

// RandomPerson returns a random person from the node list.
func (b *Bacon) RandomPerson() Node {
	// Faster than trying to pull a random person from the People map
	// Randomly probe the NodeInfo list until a person is found.
	for {
		i := rand.Intn(len(b.NodeInfo))
		if b.NodeInfo[i].Kind == "person" {
			return Node(i)
		}
	}
}

func NewPath(source Node, dest Node, parents []Node, b *Bacon) Path {
	path := make(Path, 0, 10)
	path = append(path, b.NodeInfo[dest])
	nextNode := parents[dest]
	for nextNode != source {
		path = append(path, b.NodeInfo[nextNode])
		nextNode = parents[nextNode]
	}
	path = append(path, b.NodeInfo[source])
	return path
}

// Degrees returns the Bacon Number of a given Path
func (p Path) Degrees() int {
	return (len(p) - 1) / 2
}

// Prose converts a path into readable text
func (p Path) Prose() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s and %s are separated by %d degrees. ", p[0].Name, p[len(p)-1].Name, p.Degrees())
	fmt.Fprintf(&sb, "%s was in %s with %s", p[0].Name, p[1].Name, p[2].Name)
	for i := 3; i < len(p); i += 2 {
		fmt.Fprintf(&sb, ", who was in %s with %s", p[i].Name, p[i+1].Name)
	}
	return sb.String()
}
