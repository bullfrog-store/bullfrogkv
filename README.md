# bullfrogkv
A distributed KV based on etcd-raft and pebble.

## Getting Started
### Compilation
It's extremely easy to compile BullfrogKV, just input the following command in the **project home directory**:
```shell
go build main.go
```
And then you will get a binary file called `main`.

### Configure Project
You can find an example configuration file `config_example.toml` under the `${BULLFROG_HOME}/config`:
```toml
[common]
node_count = 3

[route]
serve_addr = "127.0.0.1:8080"
grpc_addrs = ["127.0.0.1:6060", "127.0.0.1:6061", "127.0.0.1:6062"]

[store]
store_id = 1

[raft]
```
As you see, the number of the configuration file **must be the same** with the node count.
The more configuration entries can you see in the `${BULLFROG_HOME}/config/config.go`.

### Start Project
The most easy way for you to start the BullfrogKV is that can start 3 nodes by 3 configuration files `node1.toml/node2.toml/node3.toml` under the `${BULLFROG_HOME}/nodes`. You just need to open 3 terminals and go to the **project home directory**. Finally, input the following commands:
```shell
# in terminal 1
./main -conf=./nodes/node1.toml

# in terminal 2
./main -conf=./nodes/node2.toml

# in terminal 3
./main -conf=./nodes/node2.toml
```
You will clearly see 3 servers are up.

Now, you can **SET key-value pair**/**GET key**/**DELETE key** by HTTP command. For example, you want to SET a key-value pair `<hello, world>`:
```shell
curl '127.0.0.1:8080/set?key=hello&value=world'
```
After that, you can GET the corresponding value of the key `hello` by requesting other servers:
```shell
curl '127.0.0.1:8081/get?key=hello'
```
And you will get the correct result `world`.
