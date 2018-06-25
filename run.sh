#!/bin/bash

source .env
export $(cut -d= -f1 .env)
go run cmd/main.go
