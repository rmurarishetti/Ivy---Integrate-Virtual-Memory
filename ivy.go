package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const TOTAL_NODES int = 10
const TOTAL_DOCS int = 10

type CentralManager struct {
	nodes       map[int]*Node
	cmWaitGroup *sync.WaitGroup
	pgOwner     map[int]int
	pgCopies    map[int][]int
	msgReq      chan Message
	msgRes      chan Message
}

type Node struct {
	id            int
	cm            *CentralManager
	nodes         map[int]*Node
	nodeWaitGroup *sync.WaitGroup
	pgAccess      map[int]Permission
	pgContent     map[int]string
	writeToPg     string
	msgReq        chan Message
	msgRes        chan Message
}

type Message struct {
	senderId    int
	requesterId int
	msgType     MessageType
	page        int
	content     string
}

type MessageType int

const (
	//Node to Central Manager Message Types
	READREQ MessageType = iota
	WRITEREQ
	READACK
	WRITEACK
	INVALDIATEACK
	//Central Manager to Node Message Types
	READFWD
	WRITEFWD
	INVALIDATE
	READOWNERNIL
	WRITEOWNERNIL
	//Node to Node
	READPG
	WRITEPG
)

type Permission int

const (
	READONLY Permission = iota
	READWRITE
)

func (m MessageType) String() string {
	return [...]string{
		"READREQ",
		"WRITEREQ",
		"READACK",
		"WRITEACK",
		"INVALDIATEACK",
		"READFWD",
		"WRITEFWD",
		"INVALIDATE",
		"READOWNERNIL",
		"WRITEOWNERNIL",
		"READPG",
		"WRITEPG",
	}[m]
}

func (p Permission) String() string {
	return [...]string{
		"READONLY",
		"READWRITE",
	}[p]
}

func inArray(id int, array []int) bool {
	for _, item := range array {
		if item == id {
			return true
		}
	}
	return false
}

func createMessage(msgType MessageType, senderId int, requesterId int, page int, content string) *Message {
	msg := Message{
		msgType:     msgType,
		senderId:    senderId,
		requesterId: requesterId,
		page:        page,
		content:     content,
	}

	return &msg
}

func (cm *CentralManager) sendMessage(msg Message, recieverId int) {
	fmt.Printf("> [CM] Sending Message of type %s to Node %d\n", msg.msgType, recieverId)
	networkDelay := rand.Intn(300)
	time.Sleep(time.Millisecond * time.Duration(networkDelay))

	recieverNode := cm.nodes[recieverId]
	if msg.msgType == READOWNERNIL || msg.msgType == WRITEOWNERNIL {
		recieverNode.msgRes <- msg
	} else {
		recieverNode.msgReq <- msg
	}
}

func (cm *CentralManager) handleReadReq(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	_, exists := cm.pgOwner[page]
	if !exists {
		replyMsg := createMessage(READOWNERNIL, 0, requesterId, page, "")
		go cm.sendMessage(*replyMsg, requesterId)
		responseMsg := <-cm.msgRes
		fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", responseMsg.msgType, responseMsg.senderId)
		cm.cmWaitGroup.Done()
		return
	}

	pgOwner := cm.pgOwner[page]
	pgCopySet := cm.pgCopies[page]

	replyMsg := createMessage(READFWD, 0, requesterId, page, "")
	go cm.sendMessage(*replyMsg, pgOwner)
	responseMsg := <-cm.msgRes
	fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", responseMsg.msgType, responseMsg.senderId)
	if !inArray(requesterId, pgCopySet) {
		pgCopySet = append(pgCopySet, requesterId)
	}
	cm.pgCopies[page] = pgCopySet
	cm.cmWaitGroup.Done()
}

func (cm *CentralManager) handleWriteReq(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	_, exists := cm.pgOwner[page]
	if !exists {
		cm.pgOwner[page] = requesterId
		replyMsg := createMessage(WRITEOWNERNIL, 0, requesterId, page, "")
		go cm.sendMessage(*replyMsg, requesterId)
		responseMsg := <-cm.msgRes
		fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", responseMsg.msgType, responseMsg.senderId)
		cm.cmWaitGroup.Done()
		return
	}

	pgOwner := cm.pgOwner[page]
	pgCopySet := cm.pgCopies[page]

	invalidationMsg := createMessage(INVALIDATE, 0, requesterId, page, "")
	invalidationMsgCount := len(pgCopySet)

	for _, nodeid := range pgCopySet {
		go cm.sendMessage(*invalidationMsg, nodeid)
	}

	for i := 0; i < invalidationMsgCount; i++ {
		<-cm.msgRes
	}

	responseMsg := createMessage(WRITEFWD, 0, requesterId, page, "")
	go cm.sendMessage(*responseMsg, pgOwner)
	writeAckMsg := <-cm.msgRes
	fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", writeAckMsg.msgType, writeAckMsg.senderId)
	cm.pgOwner[page] = requesterId
	cm.pgCopies[page] = []int{}
	cm.cmWaitGroup.Done()
}

func (cm *CentralManager) handleIncomingMessages() {
	for {
		reqMsg := <-cm.msgReq
		fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", reqMsg.msgType, reqMsg.senderId)
		switch reqMsg.msgType {
		case READREQ:
			cm.handleReadReq(reqMsg)
		case WRITEREQ:
			cm.handleWriteReq(reqMsg)
		}
	}
}

func (node *Node) sendMessage(msg Message, recieverId int) {
	if recieverId != 0 {
		fmt.Printf("> [Node %d] Sending Message of type %s to Node %d\n", node.id, msg.msgType, recieverId)
	} else {
		fmt.Printf("> [Node %d] Sending Message of type %s to CM\n", node.id, msg.msgType)
	}
	networkDelay := rand.Intn(300)
	time.Sleep(time.Millisecond * time.Duration(networkDelay))
	if recieverId == 0 {
		if msg.msgType == READREQ || msg.msgType == WRITEREQ {
			node.cm.msgReq <- msg
		} else if msg.msgType == INVALDIATEACK || msg.msgType == READACK || msg.msgType == WRITEACK {
			node.cm.msgRes <- msg
		}
	} else {
		node.nodes[recieverId].msgRes <- msg
	}
}

func (node *Node) handleReadFwd(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	responseMsg := createMessage(READPG, node.id, requesterId, page, node.pgContent[page])
	go node.sendMessage(*responseMsg, requesterId)
}

func (node *Node) handleWriteFwd(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	responseMsg := createMessage(WRITEPG, node.id, requesterId, page, node.pgContent[page])
	delete(node.pgAccess, page)
	//delete(node.pgContent, page)
	go node.sendMessage(*responseMsg, requesterId)
}

func (node *Node) handleInvalidate(msg Message) {
	page := msg.page
	delete(node.pgAccess, page)
	//delete(node.pgContent, page)

	responseMsg := createMessage(INVALIDATE, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

func (node *Node) handleReadOwnerNil(msg Message) {
	page := msg.page
	fmt.Printf("> [Node %d] Recieved Message of type %s for Page %d\n", node.id, msg.msgType, page)
	responseMsg := createMessage(READACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

func (node *Node) handleWriteOwnerNil(msg Message) {
	page := msg.page
	fmt.Printf("> [Node %d] Recieved Message of type %s for Page %d\n", node.id, msg.msgType, page)

	node.pgContent[page] = node.writeToPg
	node.pgAccess[page] = READWRITE

	responseMsg := createMessage(WRITEACK, node.id, msg.requesterId, page, "")
	fmt.Printf("> [Node %d] Writing to Page %d\n Content:%s\n", node.id, page, node.writeToPg)
	go node.sendMessage(*responseMsg, 0)
}

func (node *Node) handleReadPg(msg Message) {
	page := msg.page
	content := msg.content

	node.pgAccess[page] = READONLY
	node.pgContent[page] = content

	fmt.Printf("> [Node %d] Recieved Page %d Content from Owner for Reading\n Content: %s\n", node.id, page, content)
	responseMsg := createMessage(READACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

func (node *Node) handleWritePg(msg Message) {
	page := msg.page
	content := msg.content

	fmt.Printf("> [Node %d] Recieved Old Page %d Content from Owner for Writing\n Content: %s\n", node.id, page, content)
	node.pgAccess[page] = READWRITE
	node.pgContent[page] = node.writeToPg
	fmt.Printf("> [Node %d] Writing to Page %d\n Content: %s\n", node.id, page, node.writeToPg)

	responseMsg := createMessage(WRITEACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}


