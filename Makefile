.PHONY: dispatcher test_game test_client runall rundispatcher rungame runclient killdispatcher killgame killclient killall

all: dispatcher test_game test_client

dispatcher:
	cd components/dispatcher && go build

test_game:
	cd examples/test_game && go build

test_client:
	cd examples/test_client && go build

rundispatcher: dispatcher
	components/dispatcher/dispatcher

rungame: test_game
	examples/test_game/test_game -sid=1

runclient: test_client
	examples/test_client/test_client -N $(N)

start:
	make all
	components/dispatcher/dispatcher &
	examples/test_game/test_game -sid=1 -log info &
	examples/test_game/test_game -sid=2 -log info &


killall:
	-make killclient
	-make killgame
	-sleep 1
	-make killdispatcher

killdispatcher:
	killall dispatcher

killgame:
	killall test_game

killclient:
	killall test_client
