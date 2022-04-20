#!/bin/bash

export SPANNER_PROJECT_ID=development-344820
export SPANNER_INSTANCE_ID=cymbal-games
export SPANNER_DATABASE_ID=my_game


wrench migrate up --directory ./schema
