start sh run_dispatcher.sh
go build
start ./test_server -configfile=../../goworld.ini -sid=1
start ./test_server -configfile=../../goworld.ini -sid=2
