from locust import HttpUser, task, events
from locust.exception import RescheduleTask

import string
import json
import random
import requests
import os

# import google.oauth2.id_token
# import google.auth.transport.requests

# request = google.auth.transport.requests.Request()

# Generate user load with 3:1 reads to write
class PlayerLoad(HttpUser):
    # def on_start(self):
    #     self.getValidUsernames()

    # def getValidUsernames(self):
    #     self.getIdToken()
    #     headers = {"Content-Type": "application/json"}
    #     r = requests.get(f"{target_audience}/players", headers=headers)

    #     global users
    #     users = json.loads(r.text)

    def generatePlayerName(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def generatePassword(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def generateEmail(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft']) + ['.com'])

    @task
    def createUser(self):
        # id_token = google.oauth2.id_token.fetch_id_token(request, target_audience)
        headers = {"Content-Type": "application/json"}
        data = {"player_name": self.generatePlayerName(), "email": self.generateEmail(), "password": self.generatePassword()}

        self.client.post("/players", data=json.dumps(data), headers=headers)


    # @task(3)
    # def readUser(self):
    #     # id_token = google.oauth2.id_token.fetch_id_token(request, target_audience)
    #     username = users[random.randint(0, len(users)-1)]['username']
    #     headers = {"Authorization": f"Bearer {id_token}",
    #        "Content-Type": "application/json"}

    #     with self.client.get(f"/user/{username}", headers=headers, catch_response=True) as response:
    #         if response.status_code == 401:
    #             self.getIdToken()
    #             raise RescheduleTask()
