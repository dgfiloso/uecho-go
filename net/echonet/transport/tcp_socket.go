// Copyright 2018 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/cybergarage/uecho-go/net/echonet/log"
	"github.com/cybergarage/uecho-go/net/echonet/protocol"
)

// A TCPSocket represents a socket for TCP.
type TCPSocket struct {
	*Socket
	Conn     *net.TCPConn
	Listener net.Listener
	readBuf  []byte
}

// NewTCPSocket returns a new TCPSocket.
func NewTCPSocket() *TCPSocket {
	sock := &TCPSocket{
		Socket:  NewSocket(),
		readBuf: make([]byte, MaxPacketSize),
	}
	return sock
}

// GetFD returns the file descriptor.
func (sock *TCPSocket) GetFD() (uintptr, error) {
	f, err := sock.Conn.File()
	if err != nil {
		return 0, err
	}
	return f.Fd(), nil

}

// Bind binds to Echonet multicast address.
func (sock *TCPSocket) Bind(ifi net.Interface, port int) error {
	err := sock.Close()
	if err != nil {
		return err
	}

	addr, err := GetInterfaceAddress(ifi)
	if err != nil {
		return err
	}

	boundAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", boundAddr)
	if err != nil {
		return err
	}

	sock.Port = port
	sock.Listener = l
	sock.Interface = ifi

	return nil
}

// Close closes the current opened socket.
func (sock *TCPSocket) Close() error {
	if sock.Conn == nil {
		return nil
	}

	err := sock.Conn.Close()
	if err != nil {
		return err
	}

	sock.Conn = nil
	sock.Listener = nil
	sock.Port = 0
	sock.Interface = net.Interface{}

	return nil
}

// Write sends the specified bytes.
func (sock *TCPSocket) Write(addr string, port int, b []byte) (int, error) {
	toAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return 0, err
	}

	// Send from no binding port

	conn, err := net.DialTCP("tcp", nil, toAddr)
	if err != nil {
		return 0, err
	}

	n, err := conn.Write(b)
	log.Trace(fmt.Sprintf(logSocketWriteFormat, conn.LocalAddr().String(), toAddr.String(), n, hex.EncodeToString(b)))
	conn.Close()

	return n, err
}

// ReadMessage reads a message from the current opened socket.
func (sock *TCPSocket) ReadMessage(clientConn net.Conn) (*protocol.Message, error) {
	if sock.Conn == nil {
		return nil, errors.New(errorSocketIsClosed)
	}

	retemoAddr := clientConn.RemoteAddr()

	reader := bufio.NewReader(clientConn)
	msg, err := protocol.NewMessageWithReader(reader)
	if err != nil {
		if sock.Conn != nil {
			log.Error(fmt.Sprintf(logSocketReadFormat, sock.Conn.LocalAddr().String(), retemoAddr, 0, ""))
		}
		return nil, err
	}

	err = msg.From.ParseString(retemoAddr.String())
	if err != nil {
		return nil, err
	}

	if msg != nil && sock.Conn != nil {
		log.Trace(fmt.Sprintf(logSocketReadFormat, sock.Conn.LocalAddr().String(), retemoAddr, msg.Size(), msg.String()))
	}

	return msg, nil
}
