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
	id          int
	nodes       map[int]*Node
	cmWaitGroup *sync.WaitGroup
	pgOwner     map[int]int
	pgCopies    map[int][]int
	msgReq      chan Message
	msgRes      chan Message
	killChan    chan int
	cmChan      chan MetaMsg
}

type Node struct {
	id            int
	cm            *CentralManager
	backup        *CentralManager
	nodes         map[int]*Node
	nodeWaitGroup *sync.WaitGroup
	pgAccess      map[int]Permission
	pgContent     map[int]string
	writeToPg     string
	msgReq        chan Message
	msgRes        chan Message
	cmKillChan    chan int
}

type Message struct {
	senderId    int
	requesterId int
	msgType     MessageType
	page        int
	content     string
}

type MetaMsg struct {
	senderId int
	nodes    map[int]*Node
	pgOwner  map[int]int
	pgCopies map[int][]int
}

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

type Permission int

const (
	READONLY Permission = iota
	READWRITE
)

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

func (cm *CentralManager) PrintState() {
	fmt.Printf("**************************************************\n  CENTRAL MANAGER STATE  \n**************************************************\n")
	for page, owner := range cm.pgOwner {
		fmt.Printf("> Page: %d, Owner: %d :: Access Type: %s , Copies: %d\n", page, owner, cm.nodes[owner].pgAccess[page], cm.pgCopies[page])
	}
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
	if !inArray(requesterId, pgCopySet) {
		pgCopySet = append(pgCopySet, requesterId)
	}
	go cm.sendMessage(*replyMsg, pgOwner)
	responseMsg := <-cm.msgRes
	fmt.Printf("> [CM] Recieved Message of type %s from Node %d\n", responseMsg.msgType, responseMsg.senderId)
	cm.pgCopies[page] = pgCopySet
	cm.cmWaitGroup.Done()
}

func main() {

}
