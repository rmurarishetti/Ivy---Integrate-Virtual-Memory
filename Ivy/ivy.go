package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const TOTAL_NODES int = 7
const TOTAL_DOCS int = 10

/*
Struct to Construct a Central Manager Instance
*/
type CentralManager struct {
	nodes       map[int]*Node
	cmWaitGroup *sync.WaitGroup
	pgOwner     map[int]int
	pgCopies    map[int][]int
	msgReq      chan Message
	msgRes      chan Message
}

/*
Struct to Construct a Node Instance
*/
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

/*
Struct to Construct a Message that's passed between Nodes and CM
*/
type Message struct {
	senderId    int
	requesterId int
	msgType     MessageType
	page        int
	content     string
}

/*
Message Type Enum to distinguish types of Messages
*/
type MessageType int

const (
	//Node to Central Manager Message Types
	READREQ MessageType = iota
	WRITEREQ
	READACK
	WRITEACK
	INVALIDATEACK
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

/*
Permission Enum for READWRITE/READONLY
*/
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
		"INVALIDATEACK",
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

/*
Function to Construct a New Node for the Ivy Protocol
*/
func NewNode(id int, cm CentralManager) *Node {
	node := Node{
		id:            id,
		cm:            &cm,
		nodes:         make(map[int]*Node),
		nodeWaitGroup: &sync.WaitGroup{},
		pgAccess:      make(map[int]Permission),
		pgContent:     make(map[int]string),
		writeToPg:     "",
		msgReq:        make(chan Message),
		msgRes:        make(chan Message),
	}

	return &node
}

/*
Function to Construct a New Central Manager for the Ivy Protocol
*/
func NewCM() *CentralManager {
	cm := CentralManager{
		nodes:       make(map[int]*Node),
		cmWaitGroup: &sync.WaitGroup{},
		pgOwner:     make(map[int]int),
		pgCopies:    make(map[int][]int),
		msgReq:      make(chan Message),
		msgRes:      make(chan Message),
	}
	return &cm
}

/*
Function to Check if a given ID is part of an Array
*/
func inArray(id int, array []int) bool {
	for _, item := range array {
		if item == id {
			return true
		}
	}
	return false
}

/*
Function to Construct a Message that's passed between Nodes and CM
*/
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

/*
Function to Print the State of a Central Manager
*/
func (cm *CentralManager) PrintState() {
	fmt.Printf("**************************************************\n  CENTRAL MANAGER STATE  \n**************************************************\n")
	for page, owner := range cm.pgOwner {
		fmt.Printf("> Page: %d, Owner: %d :: Access Type: %s , Copies: %d\n", page, owner, cm.nodes[owner].pgAccess[page], cm.pgCopies[page])
	}
}

/*
Function to send a Message that's passed between Nodes and CM
*/
func (cm *CentralManager) sendMessage(msg Message, recieverId int) {
	fmt.Printf("> [CM] Sending Message of type %s to Node %d\n", msg.msgType, recieverId)
	networkDelay := rand.Intn(50)
	time.Sleep(time.Millisecond * time.Duration(networkDelay))

	recieverNode := cm.nodes[recieverId]
	if msg.msgType == READOWNERNIL || msg.msgType == WRITEOWNERNIL {
		recieverNode.msgRes <- msg
	} else {
		recieverNode.msgReq <- msg
	}
}

/*
Function to handle Incoming Read Request Msgs at CM
*/
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
	if !inArray(requesterId, pgCopySet) {
		pgCopySet = append(pgCopySet, requesterId)
	}
	go cm.sendMessage(*replyMsg, pgOwner)
	responseMsg := <-cm.msgRes
	fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", responseMsg.msgType, responseMsg.senderId)
	cm.pgCopies[page] = pgCopySet
	cm.cmWaitGroup.Done()
}

/*
Function to handle Incoming Write Request Msgs at CM
*/
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
		msg := <-cm.msgRes
		fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", msg.msgType, msg.senderId)
	}

	responseMsg := createMessage(WRITEFWD, 0, requesterId, page, "")
	go cm.sendMessage(*responseMsg, pgOwner)
	writeAckMsg := <-cm.msgRes
	fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", writeAckMsg.msgType, writeAckMsg.senderId)
	cm.pgOwner[page] = requesterId
	cm.pgCopies[page] = []int{}
	cm.cmWaitGroup.Done()
}

/*
Function to handle Incoming Msgs at CM
*/
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

/*
Function to send messages at Node
*/
func (node *Node) sendMessage(msg Message, recieverId int) {
	if recieverId != 0 {
		fmt.Printf("> [Node %d] Sending Message of type %s to Node %d\n", node.id, msg.msgType, recieverId)
	} else {
		fmt.Printf("> [Node %d] Sending Message of type %s to CM\n", node.id, msg.msgType)
	}
	networkDelay := rand.Intn(50)
	time.Sleep(time.Millisecond * time.Duration(networkDelay))
	if recieverId == 0 {
		if msg.msgType == READREQ || msg.msgType == WRITEREQ {
			node.cm.msgReq <- msg
		} else if msg.msgType == INVALIDATEACK || msg.msgType == READACK || msg.msgType == WRITEACK {
			node.cm.msgRes <- msg
		}
	} else {
		node.nodes[recieverId].msgRes <- msg
	}
}

/*
Function to handle Read Forward Msgs at Node
*/
func (node *Node) handleReadFwd(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	fmt.Printf("> [Node %d] Current AccessType: %s for Page %d\n", node.id, node.pgAccess[page], page)
	if node.pgAccess[page] == READWRITE {
		node.pgAccess[page] = READONLY
	}

	responseMsg := createMessage(READPG, node.id, requesterId, page, node.pgContent[page])
	go node.sendMessage(*responseMsg, requesterId)
}

/*
Function to handle Write Forward Msgs at Node
*/
func (node *Node) handleWriteFwd(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	responseMsg := createMessage(WRITEPG, node.id, requesterId, page, node.pgContent[page])
	delete(node.pgAccess, page)
	//delete(node.pgContent, page)
	go node.sendMessage(*responseMsg, requesterId)
}

/*
Function to handle Invalidate Msgs at Node
*/
func (node *Node) handleInvalidate(msg Message) {
	page := msg.page
	delete(node.pgAccess, page)
	//delete(node.pgContent, page)

	responseMsg := createMessage(INVALIDATEACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

/*
Function to handle Read Owner Nil Msgs at Node
*/
func (node *Node) handleReadOwnerNil(msg Message) {
	page := msg.page
	fmt.Printf("> [Node %d] Recieved Message of type %s for Page %d\n", node.id, msg.msgType, page)
	responseMsg := createMessage(READACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

/*
Function to handle Write Owner Nil Msgs at Node
*/
func (node *Node) handleWriteOwnerNil(msg Message) {
	page := msg.page
	fmt.Printf("> [Node %d] Recieved Message of type %s for Page %d\n", node.id, msg.msgType, page)

	node.pgContent[page] = node.writeToPg
	node.pgAccess[page] = READWRITE

	responseMsg := createMessage(WRITEACK, node.id, msg.requesterId, page, "")
	fmt.Printf("> [Node %d] Writing to Page %d\n Content:%s\n", node.id, page, node.writeToPg)
	go node.sendMessage(*responseMsg, 0)
}

/*
Function to handle Read Page Msgs at Node
*/
func (node *Node) handleReadPg(msg Message) {
	page := msg.page
	content := msg.content

	node.pgAccess[page] = READONLY
	node.pgContent[page] = content

	fmt.Printf("> [Node %d] Recieved Page %d Content from Owner for Reading\n Content: %s\n", node.id, page, content)
	responseMsg := createMessage(READACK, node.id, msg.requesterId, page, "")
	go node.sendMessage(*responseMsg, 0)
}

/*
Function to handle Write Page Msgs at Node
*/
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

/*
Function to handle Incoming Msgs at Node
*/
func (node *Node) handleIncomingMessage() {
	for {
		msg := <-node.msgReq
		fmt.Printf("> [Node %d] Recieved Message of type %s from CM\n", node.id, msg.msgType)
		switch msg.msgType {
		case READFWD:
			node.handleReadFwd(msg)
		case WRITEFWD:
			node.handleWriteFwd(msg)
		case INVALIDATE:
			node.handleInvalidate(msg)
		}
	}
}

/*
Function to perform a Read End to End at Node
*/
func (node *Node) executeRead(page int) {
	node.nodeWaitGroup.Add(1)
	if _, exists := node.pgAccess[page]; exists {
		content := node.pgContent[page]
		fmt.Printf("> [Node %d] Reading Cached Page %d Content: %s\n", node.id, page, content)
		node.nodeWaitGroup.Done()
		return
	}

	readReqMsg := createMessage(READREQ, node.id, node.id, page, "")
	go node.sendMessage(*readReqMsg, 0)

	msg := <-node.msgRes
	switch msg.msgType {
	case READOWNERNIL:
		node.handleReadOwnerNil(msg)
	case READPG:
		node.handleReadPg(msg)
	}
}

/*
Function to perform a write End to End at Node
*/
func (node *Node) executeWrite(page int, content string) {
	node.nodeWaitGroup.Add(1)
	if accessType, exists := node.pgAccess[page]; exists {
		if accessType == READWRITE && node.pgContent[page] == content {
			fmt.Printf("> [Node %d] Content is same as what is trying to be written for Page %d\n", node.id, page)
			node.nodeWaitGroup.Done()
			return
		} else if accessType == READWRITE {
			node.writeToPg = content
			node.pgAccess[page] = READWRITE
			node.pgContent[page] = node.writeToPg
			fmt.Printf("> [Node %d] Writing to Page %d\n Content: %s\n", node.id, page, node.writeToPg)

			responseMsg := createMessage(WRITEACK, node.id, node.id, page, "")
			go node.sendMessage(*responseMsg, 0)
			return
		}
	}

	node.writeToPg = content
	writeReqMsg := createMessage(WRITEREQ, node.id, node.id, page, "")
	go node.sendMessage(*writeReqMsg, 0)

	msg := <-node.msgRes
	switch msg.msgType {
	case WRITEOWNERNIL:
		node.handleWriteOwnerNil(msg)
	case WRITEPG:
		node.handleWritePg(msg)
	}
}

func main() {
	var wg sync.WaitGroup

	fmt.Printf("**************************************************\n  IVY PROTOCOL (AUTOMATED NO FAULT BENCHMARK)  \n**************************************************\n")
	fmt.Printf("The network will have %d Nodes.\n", TOTAL_NODES)

	fmt.Printf("\n\nThe program will start soon....\nInstructions: The Program will be fully Automated, just watch the messages log to understand the flow. \n\n")

	cm := NewCM()
	cm.cmWaitGroup = &wg

	nodeMap := make(map[int]*Node)
	for i := 1; i <= TOTAL_NODES; i++ {
		node := NewNode(i, *cm)
		node.nodeWaitGroup = &wg
		nodeMap[i] = node
	}
	cm.nodes = nodeMap

	for _, nodei := range nodeMap {
		for _, nodej := range nodeMap {
			if nodei.id != nodej.id {
				nodei.nodes[nodej.id] = nodej
			}
		}
	}

	go cm.handleIncomingMessages()
	for _, node := range nodeMap {
		go node.handleIncomingMessage()
	}

	start := time.Now()
	for i := 1; i <= TOTAL_NODES; i++ {
		nodeMap[i].executeRead(i)
	}
	for i := 1; i <= TOTAL_NODES; i++ {
		toWrite := fmt.Sprintf("This is written by node id %d", i)
		nodeMap[i].executeWrite(i, toWrite)
	}
	for i := 1; i <= TOTAL_NODES; i++ {
		temp := i + 1
		temp %= (TOTAL_DOCS + 1)
		if temp == 0 {
			temp += 1
		}
		nodeMap[i].executeRead(temp)
	}
	for i := 1; i <= TOTAL_NODES; i++ {
		toWrite := fmt.Sprintf("This is written by pid %d", i)
		temp := i + 1
		temp %= (TOTAL_DOCS + 1)
		if temp == 0 {
			temp += 1
		}
		nodeMap[i].executeWrite(temp, toWrite)
		//nodeMap[i].executeRead((temp+1)%TOTAL_DOCS)
	}
	wg.Wait()
	end := time.Now()
	fmt.Printf("**************************************************\n CONCLUSION  \n**************************************************\n")
	cm.PrintState()
	time.Sleep(time.Second * 1)
	fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())
}
