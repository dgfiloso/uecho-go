// Copyright 2018 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"net"

	"github.com/cybergarage/uecho-go/net/echonet/log"
	"github.com/cybergarage/uecho-go/net/echonet/protocol"
)

// A MulticastHandler represents a listener for MulticastServer.
type MulticastHandler interface {
	protocol.MessageHandler
}

// A MulticastServer represents a multicast server.
type MulticastServer struct {
	*Server
	Socket        *MulticastSocket
	Handler       MulticastHandler
	UnicastServer *UnicastServer
}

// NewMulticastServer returns a new MulticastServer.
func NewMulticastServer() *MulticastServer {
	server := &MulticastServer{
		Server:        NewServer(),
		Socket:        NewMulticastSocket(),
		Handler:       nil,
		UnicastServer: nil,
	}
	return server
}

// SetHandler set a listener.
func (server *MulticastServer) SetHandler(l MulticastHandler) {
	server.Handler = l
}

// SetUnicastServer set a unicast server to response received messages.
func (server *MulticastServer) SetUnicastServer(s *UnicastServer) {
	server.UnicastServer = s
}

// Start starts this server.
func (server *MulticastServer) Start(ifi net.Interface) error {
	err := server.Socket.Bind(ifi)
	if err != nil {
		return err
	}
	server.Interface = ifi
	go handleMulticastConnection(server)
	return nil
}

// Stop stops this server.
func (server *MulticastServer) Stop() error {
	err := server.Socket.Close()
	if err != nil {
		return err
	}
	return nil
}

func handleMulticastConnection(server *MulticastServer) {
	for {
		reqMsg, err := server.Socket.ReadMessage()
		if err != nil {
			break
		}

		server.Socket.outputReadLog(log.LoggerLevelTrace, logSocketTypeUDPMulticast, reqMsg.From.String(), reqMsg.String(), reqMsg.Size())

		if server.Handler == nil {
			continue
		}

		resMsg, err := server.Handler.ProtocolMessageReceived(reqMsg)
		if server.UnicastServer != nil && err != nil && resMsg != nil {
			server.UnicastServer.ResponseMessageForRequestMessage(reqMsg, resMsg)
		}
	}
}
