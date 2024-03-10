package net

import (
	"fmt"
	"net"

	"github.com/lukejoshuapark/tcpwatch/ui"
)

const BufSize = 8192

func Begin(ui *ui.UI, localPort uint16, remoteHost string, remotePort uint16) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", localPort))
	if err != nil {
		return err
	}
	defer l.Close()

	ui.WriteLog("[grey::-]Listening on [green::b]Port %v[grey::-]...\n", localPort)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		pid := ui.AddPage()
		go setupConn(ui, pid, conn, remoteHost, remotePort)
	}
}

func setupConn(ui *ui.UI, pid int, inConn net.Conn, remoteHost string, remotePort uint16) {
	ui.WriteLog("[grey::-]Accepted connection [blue::b]%v[grey::-], initiating new connection to [yellow::b]%v:%v[grey::-]\n", pid, remoteHost, remotePort)

	outConn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", remoteHost, remotePort))
	if err != nil {
		ui.SetPageClosed(pid)
		ui.WriteLog("[grey::-]Failed to connect connection [blue::b]%v[grey::-] to remote host: [red::-]%v[grey::-]\n", pid, err)
		return
	}

	ui.WriteLog("[grey::-]Successful proxy connection to remote host for connection [blue::b]%v[grey::-]\n", pid)

	ui.SetPageConnected(pid)

	go handleConn(true, inConn, outConn, ui, pid)
	go handleConn(false, outConn, inConn, ui, pid)
}

func handleConn(isClient bool, inConn, outConn net.Conn, ui *ui.UI, pid int) {
	buf := make([]byte, BufSize)

	for {
		n, err := inConn.Read(buf)
		if err != nil {
			inConn.Close()
			outConn.Close()
			ui.SetPageClosed(pid)

			if isClient {
				ui.WriteLog("[grey::-]Connection [blue::b]%v[grey::-] terminated\n", pid)
			}

			return
		}

		if isClient {
			ui.AddClientData(pid, buf[:n])
		} else {
			ui.AddServerData(pid, buf[:n])
		}

		if _, err := outConn.Write(buf[:n]); err != nil {
			inConn.Close()
			outConn.Close()
			ui.SetPageClosed(pid)
			return
		}
	}
}
