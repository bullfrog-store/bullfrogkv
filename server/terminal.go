package server

import (
	"bullfrogkv/logger"
	"fmt"
	"strings"
)

const (
	commandSet    = "set"
	commandGet    = "get"
	commandDelete = "delete"
	commandHelp   = "help"
)

type Tml struct {
	srv *BullfrogServer
}

func Terminal(srv *BullfrogServer) *Tml {
	return &Tml{srv}
}

func (t *Tml) ScanAndProcessCommand() {
	for {
		var err error
		var args [3]string
		if _, err = fmt.Scanf("%s", &args[0]); err != nil {
			logger.Errorf("Encounter an unexpected error when input the command type (%s), error: %s", args[0], err.Error())
		}
		if len(args[0]) == 0 || args[0][0] == '\n' {
			continue
		}
		args[0] = strings.ToLower(args[0])

		if args[0] == commandSet {
			if _, err = fmt.Scanf("%s%s", &args[1], &args[2]); err != nil {
				logger.Errorf("Encounter an unexpected error when input the command key-value (%s-%s), error: %s", args[1], args[2], err.Error())
			}
		} else if args[0] == commandGet || args[0] == commandDelete {
			if _, err = fmt.Scanf("%s", &args[1]); err != nil {
				logger.Errorf("Encounter an unexpected error when input the command key (%s), err: %s", args[1], err.Error())
			}
		}

		switch args[0] {
		case commandSet:
			if err = t.srv.engine.Set(toBytes(args[1]), toBytes(args[2])); err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
			fmt.Println("OK")
		case commandGet:
			var value []byte
			if value, err = t.srv.engine.Get(toBytes(args[1])); err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
			fmt.Println(string(value))
		case commandDelete:
			if err = t.srv.engine.Delete(toBytes(args[1])); err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
			fmt.Println("OK")
		case commandHelp:
			fmt.Println("SET <KEY> <VALUE>: insert a key-value pair")
			fmt.Println("GET <KEY>: get the value of a specified key-value pair by <KEY>")
			fmt.Println("DELETE <KEY>: delete a key-value pair by <KEY>")
		default:
			fmt.Printf("WRONG command (%s)! Please type the command \"help\" for help\n", args[0])
		}
	}
}
