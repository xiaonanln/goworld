.PHONY: dispatcher test_server test_client runall rundispatcher runserver runclient killdispatcher killserver killclient killall

all: dispatcher test_server test_client

dispatcher:
	cd components/dispatcher && go build

test_server:
	cd examples/test_game && go build

test_client:
	cd examples/test_client && go build

rundispatcher: dispatcher
	components/dispatcher/dispatcher

runserver: test_server
	examples/test_game/test_game -sid=1

runclient: test_client
	examples/test_client/test_client -N $(N)

start:
	make all
	components/dispatcher/dispatcher &
	examples/test_game/test_game -sid=1 -log info &
	examples/test_game/test_game -sid=2 -log info &


killall:
	-make killdispatcher
	-make killserver
	-make killclient

killdispatcher:
	killall dispatcher

killserver:
	killall test_game

killclient:
	killall test_client
