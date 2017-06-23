#!/usr/bin/env bash
go build
./test_game -configfile=../../goworld.ini -sid=1 | tee server.log
# start ./test_game -configfile=../../goworld.ini -sid=2
