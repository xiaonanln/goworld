#!/usr/bin/env bash
go build
start ./test_game -configfile=../../goworld.ini -sid=1
#start ./test_game -configfile=../../goworld.ini -sid=2
