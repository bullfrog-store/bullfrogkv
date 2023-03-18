bullfrog:
	go build main.go

node1:
	./main -conf=./nodes/node1.toml

node2:
	./main -conf=./nodes/node2.toml

node3:
	./main -conf=./nodes/node3.toml

# If you want to run more nodes, just add:
# nodeN:
#     ./main -conf=./nodes/nodeN.toml
