from locust import HttpUser, task, events
from locust.exception import RescheduleTask

import string
import json
import random
import requests

# Generate games
# A game consists of 100 players. Only 1 winner randomly selected from those players
#
# Matchmaking is random list of players that are not playing
#
# To achieve this
# A locust user 'GameMatch' will start off by creating a "game"
# Then, pre-selecting a subset of users, and set a current_game attribute for those players.
# Once done, after a period of time, a winner is randomly selected.


# TODO: Matchmaking should ideally be handled by Agones. Once done, Locust test would convert to testing Agones match-making
class GameMatch(HttpUser):

    @task
    def createGame(self):
        headers = {"Content-Type": "application/json"}

        # Create the game, keep track of game_id
        # TODO: Make this configurable
        # data = {"numPlayers": 10}
        res = self.client.post("/games/create", headers=headers)

        # Close game
        data = {"gameUUID": res.text.replace('"', '')}
        self.client.put("/games/close", data=json.dumps(data), headers=headers)


