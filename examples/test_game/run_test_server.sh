start sh run_dispatcher.sh
go build
start ./test_game -configfile=../../goworld.ini -gid=1
start ./test_game -configfile=../../goworld.ini -gid=2
