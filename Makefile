.PHONY: dispatcher test_game test_client gate chatroom_demo unity_demo
.PHONY: runtestserver killtestserver test covertest install-deps

all: dispatcher test_game test_client gate chatroom_demo unity_demo

dispatcher:
	cd components/dispatcher && go build

gate:
	cd components/gate && go build

test_game:
	cd examples/test_game && go build

test_client:
	cd examples/test_client && go build

chatroom_demo:
	cd examples/chatroom_demo && go build

unity_demo:
	cd examples/unity_demo && go build

runtestserver: dispatcher gate test_game
	components/dispatcher/dispatcher &
	examples/test_game/test_game -gid=1 -log info &
	examples/test_game/test_game -gid=2 -log info &
	components/gate/gate -gid 1 -log debug &
	components/gate/gate -gid 2 -log debug &

killtestserver:
	- killall gate
	- sleep 3
	- killall test_game
	- sleep 5
	- killall dispatcher

test:
	go test -v `go list ./... | grep -v "/vendor/"`

covertest:
	go test -v -covermode=count `go list ./... | grep -v "/vendor/"`

install-deps:
	pip install psutil
	dep ensure
