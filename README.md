# goworld
Server Server Engine in Golang for MMORPGs

**Download goworld:**

`go get github.com/xiaonanln/goworld`

**Run goworld example server:**
1. Get MongoDB or Redis Running
2. Copy goworld.ini.sample to goworld.ini, and configure accordingly `cp goworld.ini.sample goworld.ini`
3. Build and run dispatcher: `cd components/dispatcher; go build; ./dispatcher`
4. Build and run test_game: `cd examples/test_game; go build; ./test_game -sid 1`
5. Build and run gate: `cd components/gate; go build; ./gate -gid 1`
6. Build and run test_client: `cd examples/test_client; go build; ./test_client -N 500`

**中文站点：`http://goworldgs.com/?p=64`**
