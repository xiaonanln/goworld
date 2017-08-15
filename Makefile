.PHONY: dispatcher test_game test_client gate chatroom_demo
.PHONY: runtestserver killtestserver test covertest install-deps

all: dispatcher test_game test_client gate chatroom_demo

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
	go test ./...

covertest:
	go test -covermode=count ./...

install-deps:
	- pip install psutil
	- go get golang.org/x/net/context
	- go get golang.org/x/net/websocket
	- go get github.com/xiaonanln/go-xnsyncutil/xnsyncutil
	- go get github.com/xiaonanln/goTimer
	- go get github.com/xiaonanln/typeconv
	- go get github.com/Sirupsen/logrus
	- go get github.com/garyburd/redigo/redis
	- go get github.com/google/btree
	- go get github.com/pkg/errors
	- go get github.com/bmizerany/assert
	- go get github.com/go-ini/ini
	- go get gopkg.in/mgo.v2
	- go get github.com/vmihailenco/msgpack
	- go get github.com/natefinch/lumberjack
