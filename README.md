# Notes

Project Goal:
* "Alexa, link Kevin Bacon to Anthony Hopkins"

Baby steps:

* Get cayley installed
* Read the manual
* Figure out the right source for the data and get it installed
* Figure out the bacon number query with a REPL (possibly with the built in data)
* Get the same thing working from a custom app with cayley as a library
* Package as a lambda
  * for API gateway?
  * for Alexa


  $ go get github.com/cayleygraph/cayley
  $ cd $GOPATH/src/github.com/cayleygraph/cayley/data

Interestingly 30k movies worth of data is 3.8M zipped, good proof of concept.

Homebrew has the latest binary, `brew install cayley` is fine.

From the quickstart as application:

  $ cayley http -i ./data/30kmoviedata.nq.gz -d memstore --host=:64210

  $ g.Vertex().GetLimit(5)
  $ g.V("Kevin Bacon").All()
  $ g.V("Kevin Bacon").In("<name>").All()

The full documentation is in [the gizmo API](https://github.com/cayleygraph/cayley/blob/master/docs/GizmoAPI.md)

This documentation is clearly written by someone deeply familiar with the domain objects!

## Graph Basics

Extracted from [the discourse](https://discourse.cayley.io/t/beginners-guide-to-schema-design-working-thread/436/11).

* A graph is made up of *vertices* and *edges*. A vertex is an entity; an edge is a relationship between two entities.
* A Triple specifies an edge between two vertices.

Bob and Samantha are vertices. "Bob" "knows" "Samantha" is a triple.
Every type of triple has it's own name: <subject> <predicate> <object>.
Groups of triples can describe any graph.

RDF is a generic way of describing graphs. [w3c standard](https://www.w3.org/TR/rdf11-concepts/#data-model)

RDF is a conceptual framework, not a syntax. There are syntaxes that map onto this idea.
Turtle, JSON-LD, N-Triples.

[N-Triples](https://www.w3.org/TR/n-triples/)

URL's are in angle brackets. (IRI's, technically, which allow more characters to be used).

Literals can also be used, and are quoted.

  <http://example.org/show/218> <http://www.w3.org/2000/01/rdf-schema#label> "That Seventies Show" .

Blank Nodes are curious:

  Unlike IRIs and literals, blank nodes do not identify specific resources. Statements involving blank nodes say that something with the given relationships exists, without explicitly naming it.

In N-Triples blank nodes start with underscore:

  `_:alice <http://xmlns.com/foaf/0.1/knows> _:bob .`

Example:

  <Bob> <knows> <Samantha>
  <Samantha> <knows> <Bob> .
  <Bob> <isTheSpouseOf> <Carolyn> .

Interestingly, predicates are nodes too.

  <knows> <hasDescription> "Indicates that a person knows anther person."

Ok, so that's the actual graph.
What is a schema? A collection of types.

* Types of subjects
* Types of predicates
* Types of objects

RDF Schemas can also be known as vocabularies, languages, or ontologies.

There are some basic classes you can bring online from [RDF Schema](https://www.w3.org/TR/rdf-schema/).

* `rdf:Property`: somethign that relates a subject to an objects
* `rdf:type`: is used to indicate the subject of a triple is a class of the object of a triple
* `rdfs:label`: A human readable description of a subject


... Bummer. There is no good way to find the shortest path.
[Cayley Issue](https://github.com/cayleygraph/cayley/issues/388)


Ok. Plan B.

Go has a nice graph package!
https://godoc.org/github.com/soniakeys/graph
https://github.com/soniakeys/graph/blob/master/tutorials/dijkstra.adoc


NI == Node Index // Node Int
LI == Label Index // Label Int

This could work; but it's a BFS-shaped problem, most likely.

BFS in Go: [high performance](https://medium.com/@egonelbre/a-tale-of-bfs-4ea1b8ab5eeb)
API key (v4 read-only) is in 1Password

  curl --request GET \
    --url 'https://api.themoviedb.org/4/list/1' \
    --header 'Authorization: Bearer {access_token}' \
    --header 'Content-Type: application/json;charset=utf-8'

"40 requests every 10 seconds".
"If you exceed the limit, you will receive a 429 HTTP status with a Retry-After header. As soon your cool down period expires, you are free to continue making requests."

Library:
https://pypi.org/project/tmdbsimple/

Daily File Exports: [File Exports](https://developers.themoviedb.org/3/getting-started/daily-file-exports)

Looks like I want:

http://files.tmdb.org/p/exports/movie_ids_04_28_2017.json.gz
person_ids

The movie data looks like:

  {"adult":false,"id":100,"original_title":"Lock, Stock and Two Smoking Barrels","popularity":5.811778,"video":false}

One idea would be to pick the top Nk movies and fetch their casts.

https://developers.themoviedb.org/3/movies/get-movie-credits

/movie/{movie_id}/credits

Sweet. Done.

So now the data wanted is:

  * get movie from the popular file (including name)
  * open file in movie_json/{id}.json
  * for each data['cast']{id,name}, add a link between ($actor_id <-> $movie_id)


Great, back to the real problem, how to store this stuff.

To store on disk I can use protobuf:
https://github.com/gogo/protobuf

or [mmap](https://github.com/egonelbre/a-tale-of-bfs/blob/master/graph/loaddat.go) [code](https://github.com/edsrzf/mmap-go)

The actual graph structure is a [compact adjacency list](https://www.khanacademy.org/computing/computer-science/algorithms/graph-representation/a/representing-graphs)

I'll want a single (uint32, contiguous) space for my node ids.
So I should build out that as the input to this graph:

id, type: (movie, person), external_id: ($movieid or $personid), Actual string

And I'll want to build the adjacency list.

For each movie:
  Assign ID
  For each person in movie:
    If doesn't already have an ID, assign one
    Add movie to person adjacency list
    Add person to movie adjacency list

# Searching For Golang

Swell. A graph with 10k movies, with all the data I need, is 13M uncompressed JSON. Not a crazy starting point.

Next steps:
* Get some go code written
* Define the types for what's in that file
* Import it
* Dump a random id information for fun

# Fuzzy matching

Person popularity might be useful for fuzzy matching. "If there is no match find me the most popular person with a short edit distance"

# Persistence Layer Notes

For persistence, I have a few use cases.

- Storing the graph itself
- Storing the metadata about each node (indexed by id)
- Storing the names of the nodes (and the index to to them)
- A way to do good fuzzy matching of strings to node name strings -- a subset of people only!

Boltdb might be a fast way to store this stuff?

https://npf.io/2014/07/intro-to-boltdb-painless-performant-persistence/
https://www.progville.com/go/bolt-embedded-db-golang/

It might not make a difference, but it's interesting to think about using packr for the binary:
[packr](https://github.com/gobuffalo/packr)

# Alexa Skill Notes

[Official Docs](https://developer.amazon.com/docs/custom-skills/host-a-custom-skill-as-an-aws-lambda-function.html)
[Terraform Example](https://github.com/terraform-providers/terraform-provider-aws/tree/master/examples/alexa)

* [Create a Skill](https://developer.amazon.com/docs/devconsole/create-a-skill-and-choose-the-interaction-model.html)
* copy the skill ID

* [Alexa Console](https://developer.amazon.com/alexa)

There's some built-in intents that seem to be handy for actors!

  AMAZON.SearchAction<object@VideoCreativeWork[actor]>

Looks like that's designed to take in other information and help you find an actor, not to get an actor to your skill.

  {
    "request": {
        "type": "IntentRequest",
        "locale": "en-US",
        "intent": {
            "name": "AMAZON.SearchAction<object@Actor>",
            "slots": {
                "object.type": {
                    "name": "object.type",
                    "value": "actor"
                }
            }
        }
    }
  }


  Slot values:
  {
    "name": "object.type",
    "value": "actors"
  }

  Open Bacon Guru
  > Welcome to Bacon Guru. If you give me 2 actor's names, I'll tell you how they are connected.
  > Who is the first actor?
  > Who is the second actor?
  > A and B are connected ...

  AMAZON.Actor


  > Ask Bacon Guru to connect {AMAZON.Actor} to {AMAZON.Actor}

  Ok, so created a custom intent called `GetBaconPath`
  It has a `fromActor` and `toActor` slots, of type `AMAZON.Actor`


  Skill ID: `amzn1.ask.skill.584cf3cb-3b35-462c-b665-057a40b3d955`

  "In the Designer section, under Add triggers, click Alexa Skills Kit to select the trigger."
  Under Configure triggers, select Enable for Skill ID verification.
Enter your skill ID in the Skill ID edit box.
Click Add.
Click Save to save the change.

aws lambda add-permission
    --function-name hello_world
    --statement-id 1
    --action lambda:InvokeFunction
    --principal alexa-appkit.amazon.com
    --event-source-token amzn1.ask.skill.xxxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

It doesn't look like the source token [is supported by terraform yet](https://github.com/terraform-providers/terraform-provider-aws/issues/2248).

  $ terraform init
  $ terraform plan -out theplan
  $ terraform apply theplan

Kaboom! Got an ARN back, too.

  arn:aws:lambda:us-west-2:136629216070:function:bacon-guru-skill
