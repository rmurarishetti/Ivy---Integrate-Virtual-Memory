package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const TOTAL_NODES int = 1
const TOTAL_DOCS int = 10

type CentralManager struct {
	id          int
	power       OfficeState
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

type OfficeState int

const (
	INCUMBENT OfficeState = iota
	OVERTHROWN
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
	fmt.Printf("**************************************************\n  CENTRAL MANAGER %d STATE \n**************************************************\n", cm.id)
	for page, owner := range cm.pgOwner {
		fmt.Printf("> Page: %d, Owner: %d :: Access Type: %s , Copies: %d\n", page, owner, cm.nodes[owner].pgAccess[page], cm.pgCopies[page])
	}
}

func (cm *CentralManager) sendMessage(msg Message, recieverId int) {
	fmt.Printf("> [CM %d] Sending Message of type %s to Node %d\n", cm.id, msg.msgType, recieverId)
	networkDelay := rand.Intn(50)
	time.Sleep(time.Millisecond * time.Duration(networkDelay))

	recieverNode := cm.nodes[recieverId]
	if msg.msgType == READOWNERNIL || msg.msgType == WRITEOWNERNIL {
		recieverNode.msgRes <- msg
	} else {
		recieverNode.msgReq <- msg
	}
}

func (cm *CentralManager) sendMetaMessage(reciever *CentralManager) {
	metaMsg := MetaMsg{
		senderId: cm.id,
		nodes:    cm.nodes,
		pgOwner:  cm.pgOwner,
		pgCopies: cm.pgCopies,
	}
	fmt.Printf("> [CM %d] Sending MetaMsg to Backup CM %d\n", cm.id, reciever.id)

	reciever.cmChan <- metaMsg
}

func (cm *CentralManager) handleReadReq(msg Message) {
	page := msg.page
	requesterId := msg.requesterId

	_, exists := cm.pgOwner[page]
	if !exists {
		replyMsg := createMessage(READOWNERNIL, 0, requesterId, page, "")
		go cm.sendMessage(*replyMsg, requesterId)
		responseMsg := <-cm.msgRes
		fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, responseMsg.msgType, responseMsg.senderId)
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
	fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, responseMsg.msgType, responseMsg.senderId)
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
		fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, responseMsg.msgType, responseMsg.senderId)
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
		fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, msg.msgType, msg.senderId)
	}

	responseMsg := createMessage(WRITEFWD, 0, requesterId, page, "")
	go cm.sendMessage(*responseMsg, pgOwner)
	writeAckMsg := <-cm.msgRes
	fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, writeAckMsg.msgType, writeAckMsg.senderId)
	cm.pgOwner[page] = requesterId
	cm.pgCopies[page] = []int{}
	cm.cmWaitGroup.Done()
}

// New Function
func (cm *CentralManager) handleMetaMsg(msg MetaMsg) {
	// Sync the Data Attributes from the incoming Meta Message
	cm.nodes = msg.nodes
	cm.pgCopies = msg.pgCopies
	cm.pgOwner = msg.pgOwner

	fmt.Printf("> [Backup CM] Synced MetaMessage from Primary CM %d\n", msg.senderId)
}

func (cm *CentralManager) handleIncomingMessages() {
	for {
		select {
		case reqMsg := <-cm.msgReq:
			cm.power = INCUMBENT
			fmt.Printf("> [CM %d] Recieved Message of type %s from Node %d\n", cm.id, reqMsg.msgType, reqMsg.senderId)
			switch reqMsg.msgType {
			case READREQ:
				cm.handleReadReq(reqMsg)
			case WRITEREQ:
				cm.handleWriteReq(reqMsg)
			}
		case metaMsg := <-cm.cmChan:
			//write code to handle
			cm.power = OVERTHROWN
			fmt.Printf("> [Backup CM %d] Recieved MetaMessage from Primary CM %d\n", cm.id, metaMsg.senderId)
			cm.handleMetaMsg(metaMsg)
		case <-cm.killChan:
			cm.power = OVERTHROWN
			cm.PrintState()
			return
		}
	}
}

func (node *Node) sendMessage(msg Message, recieverId int) {
	if recieverId != 0 {
		fmt.Printf("> [Node %d] Sending Message of type %s to Node %d\n", node.id, msg.msgType, recieverId)
	} else {
		fmt.Printf("> [Node %d] Sending Message of type %s to CM %d\n", node.id, msg.msgType, node.cm.id)
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

	responseMsg := createMessage(INVALIDATEACK, node.id, msg.requesterId, page, "")
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
	fmt.Printf("> [Node %d] Writing to Page %d\n> Content:%s\n", node.id, page, node.writeToPg)
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

//handle killing func

// Edit
func (node *Node) handleIncomingMessage() {
	for {
		select {
		case msg := <-node.msgReq:
			fmt.Printf("> [Node %d] Recieved Message of type %s from CM\n", node.id, msg.msgType)
			switch msg.msgType {
			case READFWD:
				node.handleReadFwd(msg)
			case WRITEFWD:
				node.handleWriteFwd(msg)
			case INVALIDATE:
				node.handleInvalidate(msg)
			}

		case <-node.cmKillChan:
			//handle killing of cm, swap primary and backup with each other
			fmt.Printf("> [Node %d] has been notified of the Supreme Leader %d's death\n", node.id, node.cm.id)
			temp := node.backup
			node.backup = node.cm
			node.cm = temp
			fmt.Printf("> [Node %d] has accepted the new Supreme Leader %d\n", node.id, node.cm.id)
		}

	}
}

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

func NewNode(id int, cm CentralManager, backup CentralManager) *Node {
	node := Node{
		id:            id,
		cm:            &cm,
		backup:        &backup,
		nodes:         make(map[int]*Node),
		nodeWaitGroup: &sync.WaitGroup{},
		pgAccess:      make(map[int]Permission),
		pgContent:     make(map[int]string),
		writeToPg:     "",
		msgReq:        make(chan Message),
		msgRes:        make(chan Message),
		cmKillChan:    make(chan int),
	}

	return &node
}

func NewCM(id int, power OfficeState) *CentralManager {
	cm := CentralManager{
		id:          id,
		power:       power,
		nodes:       make(map[int]*Node),
		cmWaitGroup: &sync.WaitGroup{},
		pgOwner:     make(map[int]int),
		pgCopies:    make(map[int][]int),
		msgReq:      make(chan Message),
		msgRes:      make(chan Message),
		killChan:    make(chan int),
		cmChan:      make(chan MetaMsg),
	}
	return &cm
}

func (cm *CentralManager) periodicFunction(reciever *CentralManager) {
	for {
		// Your periodic task goes here
		//fmt.Println("Executing periodic task...")
		if cm.power == INCUMBENT {
			cm.sendMetaMessage(reciever)
		} else {
			fmt.Printf("> [CM %d] Waiting for MetaMsg from Incumbent CM %d\n", cm.id, reciever.id)
		}

		time.Sleep(100 * time.Millisecond) // Adjust the duration as needed
	}
}

func baselineBenchmark(nodeMap map[int]*Node, cm CentralManager, backupCM CentralManager, wg sync.WaitGroup) {

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
	}
	cm.PrintState()
	wg.Wait()
	end := time.Now()
	fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())
}

func main() {
	var wg sync.WaitGroup

	fmt.Printf("**************************************************\n FAULT TOLERANT IVY PROTOCOL  \n**************************************************\n")
	fmt.Printf("The network will have %d Nodes.\n", TOTAL_NODES)

	fmt.Printf("\n\nThe program will start soon....\nInstructions:\n\nType 1 to Simulate BASELINE FAULT FREE BENCHMARK\nor 2 to Simulate 2 PRIMARY CM FAULT (DEAD) BENCHMARK\nor 3 to PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK\nor 4 to Simulate a MULTIPLE PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK\nor Type EXIT to exit:\n\n")

	cm := NewCM(0, INCUMBENT)
	cm.cmWaitGroup = &wg

	backupCM := NewCM(1, OVERTHROWN)
	backupCM.cmWaitGroup = &wg

	nodeMap := make(map[int]*Node)
	for i := 1; i <= TOTAL_NODES; i++ {
		node := NewNode(i, *cm, *backupCM)
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
	go backupCM.handleIncomingMessages()

	for _, node := range nodeMap {
		go node.handleIncomingMessage()
	}

	go backupCM.periodicFunction(cm)
	go cm.periodicFunction(backupCM)


	
	
	time.Sleep(2 * time.Second)
	random := ""
	for {
		
		fmt.Scanf("%s", &random)

		if random == "1" {

			fmt.Printf("**************************************************\n BASELINE FAULT FREE BENCHMARK  \n**************************************************\n")
			go baselineBenchmark(nodeMap, *cm, *backupCM, wg)
			time.Sleep(time.Duration(1) * time.Second)
			wg.Wait()
			os.Exit(0)
		}

		if random == "2" {
			fmt.Printf("**************************************************\n PRIMARY CM FAULT (DEAD) BENCHMARK  \n**************************************************\n")
			go func() {

				// Your main program logic goes here
				start := time.Now()
				// TESTING
				for i := 1; i <= TOTAL_NODES; i++ {
					nodeMap[i].executeRead(i)
				}
				for i := 1; i <= TOTAL_NODES; i++ {
					toWrite := fmt.Sprintf("This is written by node id %d", i)
					nodeMap[i].executeWrite(i, toWrite)
				}
				fmt.Printf("**************************************************\n KILLING PRIMARY 0  \n**************************************************\n")
				cm.killChan <- 1
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make CM Realise its no longer Incumbent
				go cm.periodicFunction(backupCM)
				// Make Dead CM relive by listening to msgs again
				go cm.handleIncomingMessages()
		
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
				}
				wg.Wait()
				end := time.Now()
				time.Sleep(time.Duration(1) * time.Second)
				fmt.Printf("**************************************************\n CONCLUSION  \n**************************************************\n")
				cm.PrintState()
				backupCM.PrintState()
				fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())
				os.Exit(0)
			}()
		}

		if random == "3" {
			fmt.Printf("**************************************************\n PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK  \n**************************************************\n")
			go func() {

				// Your main program logic goes here
				start := time.Now()
				// TESTING
				for i := 1; i <= TOTAL_NODES; i++ {
					nodeMap[i].executeRead(i)
				}
				for i := 1; i <= TOTAL_NODES; i++ {
					toWrite := fmt.Sprintf("This is written by node id %d", i)
					nodeMap[i].executeWrite(i, toWrite)
				}
				fmt.Printf("**************************************************\n KILLING PRIMARY 0  \n**************************************************\n")
				cm.killChan <- 1
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make CM Realise its no longer Incumbent
				go cm.periodicFunction(backupCM)
				// Make Dead CM relive by listening to msgs again
				go cm.handleIncomingMessages()
	
				fmt.Printf("**************************************************\n KILLING PRIMARY 1  \n**************************************************\n")
				backupCM.killChan <- 1
				cm.PrintState()
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make Backup CM Realise its no longer Incumbent
				go backupCM.periodicFunction(cm)
				// Make Dead Backup CM relive by listening to msgs again
				go backupCM.handleIncomingMessages()
		
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
				}
				wg.Wait()
				end := time.Now()
				time.Sleep(time.Duration(1) * time.Second)
				fmt.Printf("**************************************************\n CONCLUSION  \n**************************************************\n")
				cm.PrintState()
				backupCM.PrintState()
				fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())
				os.Exit(0)
			}()
		}

		if random == "4" {
			fmt.Printf("**************************************************\n MULTIPLE PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK  \n**************************************************\n")
			go func() {

				// Your main program logic goes here
				start := time.Now()
				// TESTING
				for i := 1; i <= TOTAL_NODES; i++ {
					nodeMap[i].executeRead(i)
				}
				for i := 1; i <= TOTAL_NODES; i++ {
					toWrite := fmt.Sprintf("This is written by node id %d", i)
					nodeMap[i].executeWrite(i, toWrite)
				}
				fmt.Printf("**************************************************\n KILLING PRIMARY 0  \n**************************************************\n")
				cm.killChan <- 1
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make CM Realise its no longer Incumbent
				go cm.periodicFunction(backupCM)
				// Make Dead CM relive by listening to msgs again
				go cm.handleIncomingMessages()
	
				fmt.Printf("**************************************************\n KILLING PRIMARY 1  \n**************************************************\n")
				backupCM.killChan <- 1
				cm.PrintState()
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make Backup CM Realise its no longer Incumbent
				go backupCM.periodicFunction(cm)
				// Make Dead Backup CM relive by listening to msgs again
				go backupCM.handleIncomingMessages()
		
				for i := 1; i <= TOTAL_NODES; i++ {
					temp := i + 1
					temp %= (TOTAL_DOCS + 1)
					if temp == 0 {
						temp += 1
					}
					nodeMap[i].executeRead(temp)
				}

				fmt.Printf("**************************************************\n KILLING PRIMARY 0 AGAIN  \n**************************************************\n")
				cm.killChan <- 1
				for _, node := range nodeMap {
					node.cmKillChan <- 1
				}
				time.Sleep(100 * time.Millisecond)
				// Make CM Realise its no longer Incumbent
				go cm.periodicFunction(backupCM)
				// Make Dead CM relive by listening to msgs again
				go cm.handleIncomingMessages()


				for i := 1; i <= TOTAL_NODES; i++ {
					toWrite := fmt.Sprintf("This is written by pid %d", i)
					temp := i + 1
					temp %= (TOTAL_DOCS + 1)
					if temp == 0 {
						temp += 1
					}
					nodeMap[i].executeWrite(temp, toWrite)
				}
				wg.Wait()
				end := time.Now()
				time.Sleep(time.Duration(1) * time.Second)
				fmt.Printf("**************************************************\n CONCLUSION  \n**************************************************\n")
				cm.PrintState()
				backupCM.PrintState()
				fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())
				os.Exit(0)
			}()
	
		}

		// if random == "5" {
		// 	silentNodeDeparture(clientMap)
		// 	time.Sleep(time.Duration(TIMEOUT*(NUMBER_OF_CLIENTS-1)+1) * time.Second)

		// }

		if random == "EXIT" {
			os.Exit(0)
		}
	}
	fmt.Scanf("%s", &random)
	//cm.PrintState()
	//wg.Wait()
	//end := time.Now()
	//cm.PrintState()
	//time.Sleep(time.Second * 1)
	//fmt.Printf("Time taken = %.2f seconds \n", end.Sub(start).Seconds())

}

	// go func() {

	// 	// Your main program logic goes here

	// 	// TESTING
	// 	nodeMap[2].executeRead(3)
	// 	nodeMap[1].executeWrite(3, "This is written by pid 1")
	// 	nodeMap[3].executeRead(3)
	// 	nodeMap[2].executeRead(3)
	// 	nodeMap[4].executeWrite(3, "This is written by pid 4")
	// 	fmt.Printf("**************************************************\n KILLING PRIMARY 0  \n**************************************************\n")
	// 	cm.killChan <- 1
	// 	for _, node := range nodeMap {
	// 		node.cmKillChan <- 1
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// 	// Make CM Realise its no longer Incumbent
	// 	go cm.periodicFunction(backupCM)
	// 	// Make Dead CM relive by listening to msgs again
	// 	go cm.handleIncomingMessages()

	// 	nodeMap[1].executeRead(3)
	// 	nodeMap[9].executeWrite(3, "This is written by pid 9")
	// 	nodeMap[5].executeRead(3)
	// 	fmt.Printf("**************************************************\n KILLING PRIMARY 1  \n**************************************************\n")
	// 	backupCM.killChan <- 1
	// 	cm.PrintState()
	// 	for _, node := range nodeMap {
	// 		node.cmKillChan <- 1
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// 	// Make Backup CM Realise its no longer Incumbent
	// 	go backupCM.periodicFunction(cm)
	// 	// Make Dead Backup CM relive by listening to msgs again
	// 	go backupCM.handleIncomingMessages()

	// 	nodeMap[2].executeRead(3)
	// 	nodeMap[3].executeRead(3)
	// 	nodeMap[1].executeWrite(3, "This is written by pid 1")
	// 	nodeMap[5].executeRead(3)
	// 	nodeMap[3].executeRead(3)
	// 	nodeMap[8].executeRead(3)
	// 	nodeMap[6].executeRead(3)
	// 	fmt.Println("Reached this line now")
	// 	nodeMap[9].executeRead(3)
	// 	wg.Wait()
	// 	fmt.Printf("**************************************************\n CONCLUSION  \n**************************************************\n")
	// 	cm.PrintState()
	// 	backupCM.PrintState()

	// }()