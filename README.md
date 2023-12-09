## 🚀 50.041 Distributed Systems and Computing Programming Assignment 3
### 📚 Problem 1: 
#### 📝 Implementing Ivy Protocol for Sequential Read and Write Access to a Shared File
To run the program, run the following command in the terminal:
```
cd Ivy
go run ivy.go
```

#### Instructions to run the program:

The program is fully automated, it runs the baseline benchmark on the specified number of nodes in the network. The number of nodes can be changed by changing the value of the variable ```TOTAL_NODES``` in the ```ivy.go``` file.

#### Understanding the output:
The program will output a log of the messages exchanged between the CM and the Nodes. The CM will also output the state of the system at the end of the program.

The time taken to run the program will be displayed at the end of the program.

The output of the program will look like this:

```
**************************************************
  IVY PROTOCOL (AUTOMATED NO FAULT BENCHMARK)  
**************************************************
The network will have 3 Nodes.


The program will start soon....
Instructions: The Program will be fully Automated, just watch the messages log to understand the flow. 

> [Node 1] Sending Message of type READREQ to CM
> [CM] Recieved Message of type READREQ from Node 1
> [CM] Sending Message of type READOWNERNIL to Node 1
> [Node 1] Recieved Message of type READOWNERNIL for Page 1
> [Node 2] Sending Message of type READREQ to CM
> [Node 1] Sending Message of type READACK to CM
> [CM] Recieved Message of type READACK from Node 1
.
.
.
.
.
> [CM] Recieved Message of type WRITEACK from Node 2
> [CM] Recieved Message of type WRITEREQ from Node 3
> [CM] Sending Message of type WRITEOWNERNIL to Node 3
> [Node 3] Recieved Message of type WRITEOWNERNIL for Page 4
> [Node 3] Writing to Page 4
 Content:This is written by pid 3
> [Node 3] Sending Message of type WRITEACK to CM
> [CM] Recieved Message of type WRITEACK from Node 3
**************************************************
 CONCLUSION  
**************************************************
**************************************************
  CENTRAL MANAGER STATE  
**************************************************
> Page: 1, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 2, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 3, Owner: 2 :: Access Type: READWRITE , Copies: []
> Page: 4, Owner: 3 :: Access Type: READWRITE , Copies: []
Time taken = 0.93 seconds 
```

### 📚 Problem 2: 
#### 📝 Implementing Fault Tolerant Ivy Protocol for Sequential Read and Write Access to a Shared File
To run the program, run the following command in the terminal:
```
cd Fault\ Tolerant\ Ivy/
go run fault_tolerant_ivy.go
```

#### Instructions to run the program:
The instructions to run the program are displayed when the program is run. The number of nodes can be changed by changing the value of the variable ```TOTAL_NODES``` in the ```fault_tolerant_ivy.go``` file.

```
Instructions:

Type 1 and Hit ENTER to Simulate BASELINE FAULT FREE BENCHMARK
or 2 and Hit ENTER to Simulate PRIMARY CM FAULT (DEAD) BENCHMARK
or 3 and Hit ENTER to PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK
or 4 and Hit ENTER to Simulate a MULTIPLE PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK
or 5 and Hit ENTER to Simulate a MULTIPLE PRIMARY CM AND BACKUP CM FAULT (DEAD AND RESTART) 
```

As soon as the program starts, message logs appear indicating that messages of metadata are being passed between the Primary CM and the Backup CM. This is to ensure that the Backup CM is always aware of the state of the Primary CM and is an exact replica of the Primary CM. This is done every 100ms.

You can choose to run the program in 5 different modes:
1. Baseline Benchmark on Fault Tolerant Ivy Protocol with No Faults
2. Benchmark on Fault Tolerant Ivy Protocol with one fault in Primary CM (Permanently Dead)
3. Benchmark on Fault Tolerant Ivy Protocol with one fault in Primary CM (Dead and Restart)
4. Benchmark on Fault Tolerant Ivy Protocol with multiple faults in Primary CM (Dead and Restart)
5. Benchmark on Fault Tolerant Ivy Protocol with multiple faults in Primary CM and Backup CM (Dead and Restart)

You can run these 5 scenarios by typing the corresponding number and hitting ENTER as described in the instructions above.

The scenarios correspond to the experimentation scenarios described in the specification sheet.

#### Understanding the output:
The program will output a log of the messages exchanged between the CM and the Nodes. The CM's will also output the state of the system at the end of the program.

The time taken to run the program will be displayed at the end of the program.

The output of the program will look like this:
```
**************************************************
 FAULT TOLERANT IVY PROTOCOL  
**************************************************
The network will have 3 Nodes.
The network will have 2 CMs, CM 0 is Primary and CM 1 is a Backup.


The program will start soon....
Instructions:

Type 1 and Hit ENTER to Simulate BASELINE FAULT FREE BENCHMARK
or 2 and Hit ENTER to Simulate PRIMARY CM FAULT (DEAD) BENCHMARK
or 3 and Hit ENTER to PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK
or 4 and Hit ENTER to Simulate a MULTIPLE PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK
or 5 and Hit ENTER to Simulate a MULTIPLE PRIMARY CM AND BACKUP CM FAULT (DEAD AND RESTART) BENCHMARK
or Type EXIT and Hit ENTER to exit...
> [CM 0] Sending MetaMsg to Backup CM 1
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
> [CM 1] Recieved MetaMessage from Incumbent CM 0
> [CM 1] Synced MetaMessage from Incumbent CM 0
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
> [CM 0] Sending MetaMsg to Backup CM 1
> [CM 1] Recieved MetaMessage from Incumbent CM 0
> [CM 1] Synced MetaMessage from Incumbent CM 0
> [CM 0] Sending MetaMsg to Backup CM 1
> [CM 1] Recieved MetaMessage from Incumbent CM 0
> [CM 1] Synced MetaMessage from Incumbent CM 0
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
.
.
.
**************************************************
 PRIMARY CM FAULT (DEAD AND RESTART) BENCHMARK  
**************************************************
> [Node 1] Sending Message of type READREQ to CM 0
> [CM 0] Recieved Message of type READREQ from Node 1
> [CM 0] Sending Message of type READOWNERNIL to Node 1
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
> [CM 0] Sending MetaMsg to Backup CM 1
> [CM 1] Recieved MetaMessage from Incumbent CM 0
> [CM 1] Synced MetaMessage from Incumbent CM 0
> [Node 1] Recieved Message of type READOWNERNIL for Page 1
> [Node 2] Sending Message of type READREQ to CM 0
> [Node 1] Sending Message of type READACK to CM 0
> [CM 0] Recieved Message of type READACK from Node 1
.
.
.
**************************************************
 KILLING PRIMARY CM  
**************************************************
> [Node 3] Sending Message of type WRITEACK to CM 0
> [CM 0] Recieved Message of type WRITEACK from Node 3
**************************************************
  CENTRAL MANAGER 0 STATE 
**************************************************
> [Node 2] has been notified of the CM 0's death
> [Node 2] has accepted the new CM 1 as Incumbent
> [Node 3] has been notified of the CM 0's death
> [Node 3] has accepted the new CM 1 as Incumbent
> [Node 1] has been notified of the CM 0's death
> [Node 1] has accepted the new CM 1 as Incumbent
> Page: 3, Owner: 3 :: Access Type: READWRITE , Copies: []
> Page: 1, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 2, Owner: 2 :: Access Type: READWRITE , Copies: []
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
> [CM 0] Waiting for MetaMsg from Incumbent CM 1
.
.
.
.
> [CM 1] Recieved MetaMessage from Incumbent CM 0
> [CM 1] Synced MetaMessage from Incumbent CM 0
> [CM 1] Waiting for MetaMsg from Incumbent CM 0
**************************************************
 CONCLUSION  
**************************************************
**************************************************
  CENTRAL MANAGER 0 STATE 
**************************************************
> Page: 1, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 2, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 3, Owner: 2 :: Access Type: READWRITE , Copies: []
> Page: 4, Owner: 3 :: Access Type: READWRITE , Copies: []
**************************************************
  CENTRAL MANAGER 1 STATE 
**************************************************
> Page: 3, Owner: 2 :: Access Type: READWRITE , Copies: []
> Page: 4, Owner: 3 :: Access Type: READWRITE , Copies: []
> Page: 1, Owner: 1 :: Access Type: READWRITE , Copies: []
> Page: 2, Owner: 1 :: Access Type: READWRITE , Copies: []
Time taken = 1.01 seconds 

```

### 📚 Problem 3:
#### 📝 Proof of Sequntial Consistency of the Ivy and Fault Tolerant Ivy Implementation

<div style="text-align: justify">
<p>

The implementation of the Ivy and Fault Tolerant Ivy use channels which aren't buffered, this ensures that the Central Manager can only process one request at a time with a delay less than that the time taken to send and recieve messages in the network. This ensures that the Central Manager processes the requests in the order they are recieved.

The Ivy and Fault Tolerant Ivy implementation also use a shared file to store the data. The data is only written to the file when the Central Manager has processed the request and the node has recieved a reply from the corresponding node who is the owner of the document. This ensures that the data is written to the file in the order the requests are processed by the Central Manager.

The acknowledgements to the requests are handled in a different pathway through a non buffered channel, which are processed at the Nodes and the Central Manager. This ensures that the acknowledgements are processed in the order they are recieved.

In the Fault Tolerant Ivy, an extra channel is created to facilitate message passing such that the Backup CM and the Primary CM are always aware of the state of each other and are exact replicas of one another through a periodic message that exchanges metadata every 100ms. This ensures that the Backup CM can take over the role of the Primary CM in case of a failure and the state of the system is consistent.

In the case of a fault, the nodes are notified of the CM's failure through a channel, which triggers an instant change where the backup assumes the role of the primary. They can continue their operations according to the previous state of the system.
</p>
</div>

### 📚 Problem 4:
#### 📝 Experimentation and Performance Analysis of the Protocols 
The performance analysis of the 2 algorithms has been done by varying the number of nodes in the network with differing scenarios as described in the specification sheet. 

- All Runs are 10 Docs.
- Each message has a random delay between 0 and 50 ms associated with to simulate real network delay.


The results can be found in the Benchmarks PDF in the root directory.

```
open Benchmarks.pdf
```

##### 📝 Experiment 1: No Faults, Multiple Read and Write Requests in both fault tolerant and non-fault tolerant Ivy.

- Corresponds to Scenario 1 and Scenario 2 in the Benchmark PDF.
1. Scenario 1: Baseline Benchmark on Vanilla Ivy Protocol with No Faults
3. Scenario 2: Baseline Benchmark on Fault Tolerant Ivy Protocol with No Faults - Red

##### 📝 Experiment 2: One Fault, Multiple Read and Write Requests in fault tolerant Ivy.

- Corresponds to Scenario 3 and Scenario 4 in the Benchmark PDF.
1. Scenario 3: Benchmark on Fault Tolerant Ivy Protocol with one fault in Primary CM (Permanently Dead)
2. Scenario 4: Benchmark on Fault Tolerant Ivy Protocol with one fault in Primary CM (Dead and Restart)

##### 📝 Experiment 3: Multiple Faults in Primary CM, Multiple Read and Write Requests in fault tolerant Ivy.

- Corresponds to Scenario 5 in the Benchmark PDF.
1. Scenario 5: Benchmark on Fault Tolerant Ivy Protocol with multiple faults in Primary CM (Dead and Restart)

##### 📝 Experiment 4: Multiple Faults in Primary CM and Backup CM, Multiple Read and Write Requests in fault tolerant Ivy.

- Corresponds to Scenario 6 in the Benchmark PDF.
1. Scenario 6: Benchmark on Fault Tolerant Ivy Protocol with multiple faults in Primary CM and Backup CM (Dead and Restart) 


