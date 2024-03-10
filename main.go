package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/lukejoshuapark/tcpwatch/net"
	"github.com/lukejoshuapark/tcpwatch/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	if len(args) != 3 {
		return errors.New("unexpected arguments supplied.  Usage: tcpwatch <local-port> <remote-host> <remote-port>")
	}

	localPortStr := args[0]
	remoteHost := args[1]
	remotePortStr := args[2]

	localPort, err := strconv.ParseUint(localPortStr, 10, 16)
	if err != nil {
		return fmt.Errorf("failed to parse local port: %v", err)
	}

	remotePort, err := strconv.ParseUint(remotePortStr, 10, 16)
	if err != nil {
		return fmt.Errorf("failed to parse remote port: %v", err)
	}

	ui := ui.New(uint16(localPort), remoteHost, uint16(remotePort))

	go func() {
		if err := net.Begin(ui, uint16(localPort), remoteHost, uint16(remotePort)); err != nil {
			panic(err)
		}
	}()

	return ui.Run()
}
