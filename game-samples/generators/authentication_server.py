from locust import HttpUser, task, events
from locust.exception import RescheduleTask

import string
import json
import random
import requests

# Generate player load with 5:1 reads to write
class PlayerLoad(HttpUser):
    def on_start(self):
        self.getValidUUIDs()

    def getValidUUIDs(self):
        headers = {"Content-Type": "application/json"}
        r = requests.get(f"{self.host}/players", headers=headers)

        global pUUIDs
        pUUIDs = json.loads(r.text)

    def generatePlayerName(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def generatePassword(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def generateEmail(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft']) + ['.com'])

    @task
    def createPlayer(self):
        headers = {"Content-Type": "application/json"}
        data = {"player_name": self.generatePlayerName(), "email": self.generateEmail(), "password": self.generatePassword()}

        self.client.post("/players", data=json.dumps(data), headers=headers)

    @task(5)
    def getPlayer(self):
        pUUID = pUUIDs[random.randint(0, len(pUUIDs)-1)]
        headers = {"Content-Type": "application/json"}

        self.client.get(f"/players/{pUUID}", headers=headers)
