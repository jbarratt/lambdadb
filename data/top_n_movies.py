#!/usr/bin/env python3

""" Find the top N most popular films and dump to json """

import json
import operator

def main():
    popular = []
    with open('movie_ids_06_04_2018.json', 'r') as movies:
        for line in movies.readlines():
            movie = json.loads(line)
            popular.append((movie["id"], movie["popularity"]))
    popular.sort(key = operator.itemgetter(1), reverse=True)
    with open('movies_by_popularity.json', 'w') as of:
        json.dump(popular[0:40000], of)

if __name__ == '__main__':
    main()
