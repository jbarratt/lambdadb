package bacon

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/arbovm/levenshtein"
	"github.com/shawnsmithdev/zermelo/zuint32"
)

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
	Name     string
	IsPerson bool   `json:"isPerson"`
	TmdbID   uint32 `json:"tmdb_id"`
	Node     Node   `json:"node_id"`
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

// NewFromJSON is like NewBacon with a JSON path
func NewFromJSON(path string) (*Bacon, error) {
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

// NewFromGob is like NewBacon with a Gob path
func NewFromGob(path string) (*Bacon, error) {
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

// FindPerson finds a Person-typed node by name
// Does approximate matching where possible
func (b *Bacon) FindPerson(name string) Node {
	key := strings.ToLower(name)
	node, ok := b.People[key]
	if ok {
		return node
	}
	fmt.Println("Name not found, using most similar")
	searchStart := time.Now()
	minDist := math.MaxInt32
	minNode := Node(0)
	for person, node := range b.People {
		dist := levenshtein.Distance(person, key)
		if dist < minDist {
			minDist = dist
			minNode = node
		}
	}
	fmt.Printf("Fuzzy Finding took %s\n", time.Since(searchStart))
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

// FindPath returns bacon Path of source to dest
func (b *Bacon) FindPath(source Node, dest Node) (Path, error) {

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
					return NewPath(source, dest, parentNode, b)
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
		if b.NodeInfo[i].IsPerson {
			return Node(i)
		}
	}
}

// NewPath constructs a Path of NodeInfos
func NewPath(source Node, dest Node, parents []Node, b *Bacon) (Path, error) {
	maxPath := 60
	path := make(Path, 0, 10)
	path = append(path, b.NodeInfo[dest])
	nextNode := parents[dest]
	for nextNode != source && len(path) < maxPath {
		path = append(path, b.NodeInfo[nextNode])
		nextNode = parents[nextNode]
	}
	if len(path) >= maxPath {
		return path, errors.New("Path impossibly long")
	}
	path = append(path, b.NodeInfo[source])
	return path, nil
}

// Degrees returns the Bacon Number of a given Path
func (p Path) Degrees() int {
	return (len(p) - 1) / 2
}

// Prose converts a path into readable text
func (p Path) Prose() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s and %s are separated by %d degrees. ", p[0].Name, p[len(p)-1].Name, p.Degrees())
	fmt.Fprintf(&sb, "%s was in \"%s\" with %s", p[0].Name, p[1].Name, p[2].Name)
	for i := 3; i < len(p); i += 2 {
		fmt.Fprintf(&sb, ", who was in \"%s\" with %s", p[i].Name, p[i+1].Name)
	}
	return sb.String()
}
