package main

import "sync"

const TOTAL_NODES int = 10
const TOTAL_DOCS int = 10

type CentralManager struct{
	nodes map[int]*Node
	cmWaitGroup *sync.WaitGroup
	pgOwner map[int]int
	pgCopies map[int][]int
	msgReq chan Message
	msgRes chan Message
}

type Node struct {
	id int
	cm *CentralManager
	nodes map[int]*Node
	nodeWaitGroup *sync.WaitGroup
	pgAccess map[int]Permission
	pgContent map[int]string
	writeToPg string
	msgReq chan Message
	msgRes chan Message
}

type Message struct {
	senderId int
	requesterId int
	msgType MessageType
	page int
	content string
}

type MessageType int

const(
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
	//Node to Node
	READPG 
	WRITEPG 
)

type Permission int

const(
	READONLY Permission = iota
	READWRITE
)

func (m MessageType) Print() string {
	return [...]string{
		"READREQ",
		"WRITEREQ",
		"READACK",
		"WRITEACK",
		"INVALDIATEACK",
		"READFWD",
		"WRITEFWD", 
		"INVALIDATE",
		"READPG",
		"WRITEPG",
	}[m]
}

func (p Permission) Print() string {
	return [...]string{
		"READONLY",
		"READWRITE",
	}[p]
}

