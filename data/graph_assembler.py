#!/usr/bin/env python3

""" Build an adjacency list graph

- load the list of movies
- load the list of people

For each movie:
    - load the file
    - Assign ID
    For each person in movie:
        - If doesn't already have an ID, assign one
        - Add movie to person adjacency list
        - Add person to movie adjacency list

For every assigned id:
    - sort adjacency list
    - append to a master list
    - add offset to Span

"""

import json


def main():

    id_gen = IDGen()
    movie_names = load_movie_names() # tmdb_id -> movie_name
    people_names = load_people_names() # tmdb id -> people name
    node_data = {} # What do we know about a node id?
    node_neighbors = {} # Adjancency List (node id -> [neighbor node ids])
    person_to_node = {} # TMDB Person ID -> Node ID


    print("Building adjacency list")
    # This is a list of (id, popularity) tuples
    fetched_movies = json.load(open('movies_by_popularity.json', 'r'))
    for movie_id, _ in fetched_movies:
        node_id = id_gen.get_id()
        node_data[node_id] = {'type': 'movie', 'tmdb_id': movie_id, 'name': movie_names[movie_id]}

        movie_data = json.load(open(f"movie_json/{movie_id}.json"))

        for member in movie_data['cast']:
            person_id = member['id']
            p_node_id = None
            if person_id not in person_to_node:
                if person_id in people_names:
                    p_node_id = id_gen.get_id()
                    node_data[p_node_id] = {'type': 'person', 'tmdb_id': person_id, 'name': people_names[person_id]}
                    person_to_node[person_id] = p_node_id
            else:
                p_node_id = person_to_node[person_id]

            if p_node_id is not None:
                add_edge(node_neighbors, node_id, p_node_id)

    out_data = {}

    out_data['node_data'] = [node_data[x] for x in range(id_gen.next_id)]
    out_data['people'] = {node_data[x]['name']: x for x in range(id_gen.next_id) if node_data[x]['type'] == 'person'}


    # Dump the adjacency list
    g_list = []
    g_span = [None]*id_gen.next_id
    offset = 0
    for node_id in range(id_gen.next_id):
        g_span[node_id] = offset
        # Having these sorted leads to more efficient graph utilization
        if node_id not in node_neighbors:
            node_neighbors[node_id] = []
        node_neighbors[node_id].sort()
        g_list.extend(node_neighbors[node_id])
        offset += len(node_neighbors[node_id])

    out_data['graph'] = {'list': g_list, 'span': g_span}

    with open('bacon.json', 'w') as of:
        of.write(json.dumps(out_data))


def add_edge(nodes, a, b):
    """ Add an edge to the adjacency lists of both """
    for node in (a, b):
        if node not in nodes:
            nodes[node] = []

    if a not in nodes[b]:
        nodes[b].append(a)

    if b not in nodes[a]:
        nodes[a].append(b)

def load_movie_names():
    print("Loading movie names")
    movie_names = {}
    with open('movie_ids_06_04_2018.json', 'r') as movies:
        for line in movies.readlines():
            movie = json.loads(line)
            movie_names[movie["id"]] = movie["original_title"]
    return movie_names

def load_people_names():
    print("Loading people names")
    people_names = {}
    with open('person_ids_06_04_2018.json', 'r') as people:
        for line in people.readlines():
            person = json.loads(line)
            people_names[person["id"]] = person["name"]
    return people_names

class IDGen(object):
    """ ID Generation """
    def __init__(self):
        self.next_id = 0

    def get_id(self):
        id = self.next_id
        self.next_id += 1
        return id

if __name__ == '__main__':
    main()
