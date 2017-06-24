.PHONY: dispatcher test_server test_client

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
	examples/test_client/test_client 2>&1
