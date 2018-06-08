#!/usr/bin/env python3

""" Find the top N most popular films and dump to json """

import json
import requests
from tqdm import tqdm
import time
import os
from tokenbucket import TokenBucket

def main():
    api_key = os.environ['TMDB_API_KEY']
    movies = json.load(open('movies_by_popularity.json', 'r'))
    uri = "https://api.themoviedb.org/3/movie/{}/credits?api_key={}"

    bucket = TokenBucket(rate = 4, tokens=0, capacity=40)

    for movie_id, _ in tqdm(movies):
        if os.path.isfile(f"movie_json/{movie_id}.json"):
            continue
        bucket.consume(1)
        r = requests.get(uri.format(movie_id, api_key))
        if r.status_code == 200:
            with open(f"movie_json/{movie_id}.json", "w") as of:
                of.write(r.text)
        else:
            time.sleep(1) #oops maybe broke rate limit

if __name__ == '__main__':
    main()
