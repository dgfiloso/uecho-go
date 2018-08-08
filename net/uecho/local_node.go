// Copyright (C) 2018 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uecho

import (
	"fmt"

	"github.com/cybergarage/uecho-go/net/uecho/protocol"
)

// LocalNode is an instance for Echonet node.
type LocalNode struct {
	*baseNode
	*server
	lastTID          uint
	manufacturerCode uint
}

// NewLocalNode returns a new node.
func NewLocalNode() *LocalNode {
	node := &LocalNode{
		baseNode:         newBaseNode(),
		server:           newServer(),
		manufacturerCode: NodeManufacturerUnknown,
		lastTID:          TIDMin,
	}

	node.AddProfile(NewLocalNodeProfile())
	node.server.SetMessageListener(node)

	return node
}

// SetManufacturerCode sets a manufacture codes to the node.
func (node *LocalNode) SetManufacturerCode(code uint) {
	node.manufacturerCode = code
}

// GetManufacturerCode return the manufacture codes of the node.
func (node *LocalNode) GetManufacturerCode() uint {
	return node.manufacturerCode
}

// getNextTID returns a next TID.
func (node *LocalNode) getNextTID() uint {
	if TIDMax <= node.lastTID {
		node.lastTID = TIDMin
	} else {
		node.lastTID++
	}
	return node.lastTID
}

// AddDevice adds a new device into the node, and set the node profile and manufacture code.
func (node *LocalNode) AddDevice(dev *Device) error {
	err := node.baseNode.AddDevice(dev)
	if err != nil {
		return err
	}

	dev.SetManufacturerCode(node.manufacturerCode)
	dev.SetParentNode(node)

	return node.updateNodeProfile()
}

// AddProfile adds a new profile object into the node, and set the node profile and manufacture code.
func (node *LocalNode) AddProfile(prof *Profile) error {
	err := node.baseNode.AddProfile(prof)
	if err != nil {
		return err
	}

	prof.SetManufacturerCode(node.manufacturerCode)
	prof.SetParentNode(node)

	return node.updateNodeProfile()
}

// GetAddress returns the bound address.
func (node *LocalNode) GetAddress() string {
	addrs := node.server.GetBoundAddresses()
	if len(addrs) <= 0 {
		return ""
	}
	return addrs[0]
}

// GetPort returns the bound address.
func (node *LocalNode) GetPort() int {
	return node.server.GetPort()
}

// AnnounceMessage announces a message.
func (node *LocalNode) AnnounceMessage(msg *protocol.Message) error {
	msg.SetTID(node.getNextTID())
	return node.server.NotifyMessage(msg)
}

// AnnounceProperty announces a specified property.
func (node *LocalNode) AnnounceProperty(prop *Property) error {
	msg := protocol.NewMessage()
	msg.SetESV(protocol.ESVNotification)
	msg.SetSourceObjectCode(NodeProfileObjectReadOnly)
	msg.SetDestinationObjectCode(NodeProfileObject)
	msg.AddProperty(prop.toProtocolProperty())

	return node.AnnounceMessage(msg)
}

// Announce announces the node
func (node *LocalNode) Announce() error {
	nodePropObj, err := node.GetNodeProfile()
	if err != nil {
		return err
	}

	nodeProp, ok := nodePropObj.GetProperty(NodeProfileClassInstanceListNotification)
	if !ok {
		return fmt.Errorf(errorObjectProfileObjectNotFound)
	}

	return node.AnnounceProperty(nodeProp)
}

// SendMessage send a message to the node
func (node *LocalNode) SendMessage(dstNode Node, msg *protocol.Message) error {
	_, err := node.server.SendMessage(string(dstNode.GetAddress()), dstNode.GetPort(), msg)
	return err
}

// Start starts the node.
func (node *LocalNode) Start() error {
	err := node.server.Start()
	if err != nil {
		return err
	}

	err = node.Announce()
	if err != nil {
		return err
	}

	return nil
}

// Stop stop the node.
func (node *LocalNode) Stop() error {
	err := node.server.Stop()
	if err != nil {
		return err
	}

	return nil
}

// updateNodeProfile updates the node profile in the node.
func (node *LocalNode) updateNodeProfile() error {
	nodeProf, err := node.GetNodeProfile()
	if err != nil {
		return err
	}

	// Check the current all objects

	classes := make([]*Class, 0)

	for _, dev := range node.devices {
		devClass := dev.GetClass()
		hasSameClass := false
		for _, class := range classes {
			if class.Equals(devClass) {
				hasSameClass = true
				break
			}
		}
		if hasSameClass {
			continue
		}
		classes = append(classes, devClass)
	}

	for _, prof := range node.profiles {
		profClass := prof.GetClass()
		hasSameClass := false
		for _, class := range classes {
			if class.Equals(profClass) {
				hasSameClass = true
				break
			}
		}
		if hasSameClass {
			continue
		}
		classes = append(classes, profClass)
	}

	// Number of self-node instances

	instanceCount := uint(len(node.devices))
	nodeProf.SetInstanceCount(instanceCount)

	// Number of self-node classes

	nodeProf.SetClassCount(uint(len(classes)))

	// Self-node instance list S and Instance list notification

	nodeProf.SetInstanceList(node.devices)

	// Self-node class list S

	nodeProf.SetClassList(classes)

	return nil
}
