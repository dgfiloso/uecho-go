// Copyright 2017 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"
	"fmt"
	"net"
)

// A MulticastSocket represents a socket.
type MulticastSocket struct {
	*UDPSocket
}

// NewMulticastSocket returns a new MulticastSocket.
func NewMulticastSocket() *MulticastSocket {
	sock := &MulticastSocket{}
	sock.UDPSocket = NewUDPSocket()
	return sock
}

// Bind binds to Echonet multicast address.
func (sock *MulticastSocket) Bind(ifi net.Interface) error {
	err := sock.Close()
	if err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", MULTICAST_ADDRESS)
	if err != nil {
		return err
	}

	sock.Conn, err = net.ListenMulticastUDP("udp", &ifi, addr)
	if err != nil {
		return fmt.Errorf("%s (%s)", err.Error(), ifi.Name)
	}

	sock.Interface = ifi

	return nil
}

// Write sends the specified bytes.
func (sock *MulticastSocket) Write(b []byte) (int, error) {
	if sock.Conn == nil {
		return 0, errors.New(errorSocketIsClosed)
	}

	addr, err := net.ResolveUDPAddr("udp", MULTICAST_ADDRESS)
	if err != nil {
		return 0, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return conn.Write(b)
}
