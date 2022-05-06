
## Must patch locust/__init__.py monkey_patch.all() function to monkey_patch.all(thread=False)

from locust import User, task
# from locust.exception import RescheduleTask
from google.cloud import spanner

import time
import string
import random
import uuid

# TODO: implement better project/instance configuration
project_id = "development-344820"
instance_id = "cymbal-games"
database_id = "my_game"

"""
Sets up a SpannerClient for each user. This has the implication that each user will
have its own session pool.
"""
class SpannerClient:
    def __init__(self, request_event):
        self._request_event = request_event

        # Instantiate a client.
        spanner_client = spanner.Client(project_id)

        # Get a Cloud Spanner instance by ID.
        instance = spanner_client.instance(instance_id)

        # Get a Cloud Spanner database by ID.
        self._database = instance.database(database_id)

    # Executes a read and and returns number of rows modified
    def _execute_read(self, query):
        with self._database.snapshot() as snapshot:
            results = snapshot.execute_sql(query)

            return len(results)

    # Executes a write and returns number of rows modified
    def _execute_write(self, query):
        def dml(transaction):
            return transaction.execute_update(query)

        row_ct = self._database.run_in_transaction(dml)
        return row_ct

    # Determines which type of query to run (read or write), and returns the row count
    def _execute_query(self, type, query):
        return {
            'insert': self._execute_write,
            # 'update': execute_write,
            # 'delete': execute_write,
            'read': self._execute_read,
        }[type](query)

    # TODO: implement query responses better
    def __getattr__(self, name):
        def wrapper(*args, **kwargs):
            request_meta = {
                "request_type": "spanner",
                "name": name,
                "start_time": time.time(),
                "response_length": 0,
                "response": None,
                "context": {},
                "exception": None,
            }

            try:
                # Results of executing the query is the number of rows returned or modified, depending on type of query
                request_meta["response_length"] = self._execute_query(*args, **kwargs)
            except Exception as e:
                request_meta["exception"] = e

            self._request_event.fire(**request_meta)  # This is what makes the request actually get logged in Locust
            return request_meta["response"]

        return wrapper

# This class will be executed when you fire up locust
class SpannerUser(User):
    abstract = True # dont instantiate this as an actual user when running Locust

    def __init__(self, environment):
        super().__init__(environment)
        self.client = SpannerClient(request_event=environment.events.request)

class PlayerTasks(SpannerUser):
    host = "http://127.0.0.1/"

    # def on_start(self):
    #     self.getValidUUIDs()

    # def getValidUUIDs(self):
    #     r = requests.get(f"{self.host}/players", headers=headers)

    #     global pUUIDs
    #     pUUIDs = json.loads(r.text)

    def _generatePlayerName(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def _generatePassword(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=15))

    def _generateEmail(self):
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=10) + ['@'] +
            random.choices(['gmail', 'yahoo', 'microsoft', 'hotmail']) + ['.com'])

    @task
    def createPlayer(self):
        data = {
                "playerUUID": str(uuid.uuid4()),
                "player_name": self._generatePlayerName(),
                "email": self._generateEmail(),
                "password": self._generatePassword()
                }
        stmt ="""INSERT players (playerUUID, player_name, email, user_password, created, active_skinUUID) VALUES
					('{playerUUID}', '{playerName}', '{email}', '{password}', CURRENT_TIMESTAMP(), '1')"""
        self.client.get_add_player("insert",
                                stmt.format(playerUUID = data['playerUUID'],
                                            playerName = data['player_name'],
                                            email = data['email'],
                                            password = data['password'])
        )

    # @task(5)
    # def getPlayer(self):
    #     # id_token = google.oauth2.id_token.fetch_id_token(request, target_audience)
    #     pUUID = pUUIDs[random.randint(0, len(pUUIDs)-1)]
    #     headers = {"Content-Type": "application/json"}

    #     self.client.get(f"/players/{pUUID}", headers=headers)
