#!/bin/bash

source .env
export $(cut -d= -f1 .env)
./5ch_slack_bot
