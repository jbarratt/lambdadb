#!/usr/bin/env python3

""" Do a sanity check of the compact graph """

import sys
import json
import random

def main():
    # Load graph & node_data
    bacon = json.load(open('bacon.json', 'r'))
    graph = bacon['graph']
    nodes = bacon['node_data']
    # Fetch Popular movies
    popular = json.load(open('movies_by_popularity.json', 'r'))
    # Select a random one
    random_tmdb_movie = random.choice(popular)[0]

    # find our node id that matches
    node_id = None
    for id, data in enumerate(nodes):
        if data['type'] == 'movie' and data['tmdb_id'] == random_tmdb_movie:
            node_id = id
            print(f"Graph Movie Name: {data['name']}")
            break
    if node_id is None:
        print(f"Unable to find a node for ID {random_tmdb_movie}")
        sys.exit(1)

    # Load detailed data for that movie
    print("Source Cast Data:")
    movie_data = json.load(open(f"movie_json/{random_tmdb_movie}.json"))
    for member in movie_data['cast']:
        print(member["name"])

    print("-----\n")
    # Print casts from original datga
    (start, end) = (graph['span'][node_id], graph['span'][node_id+1])
    # Print cast from graph
    print("Graph cast data:")
    print(",".join([str(x) for x in graph['list'][start:end]]))
    for member in graph['list'][start:end]:
        print(f"{nodes[member]['name']}")


if __name__ == '__main__':
    main()
