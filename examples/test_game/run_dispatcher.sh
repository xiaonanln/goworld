#!/usr/bin/env bash
cd ../../components/dispatcher
go build
./dispatcher.exe -configfile=../../goworld.ini