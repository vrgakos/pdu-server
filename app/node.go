package app

import (
	pb "pdu-server/protos"
	"log"
	"time"
)

type Node struct {
	app *App
	nid string
	name string

	score *NodeScore
	imState int
	imClient chan *Client
	imMeasure chan []*pb.MeasureData
}

type NodeScore struct {
	CpuZero uint64		`json:"cpu_zero"`
	CpuFull uint64  	`json:"cpu_full"`
	VmZero uint64		`json:"vm_zero"`
	VmFull uint64		`json:"vm_full"`
	TimerZero uint64	`json:"timer_zero"`
	TimerFull uint64	`json:"timer_full"`
}

func NewNode(app *App, nid string) *Node {
	node := &Node{
		app: app,
		nid: nid,

		imState: 0,
		imClient: make(chan *Client, 2),
		imMeasure: make(chan []*pb.MeasureData, 10),
	}

	go node.initMeasure()

	return node
}

func (n *Node) ClientConnected(client *Client) {
	n.imClient <- client
}

func (n *Node) changeIMState(newState int, name string) {
	n.imState = newState
	log.Printf("Node (%s) InitMeasure state is %s", n.nid, name)
	n.app.browserGw.BroadcastNodeIMMessage(n)
}

func (n *Node) initMeasure() {
	n.changeIMState(0, "STARTED")

	// WAIT FOR BOTH CLIENTS
	var measure *Client
	var stress *Client

	measure = n.getOneClient(pb.ClientMode_MEASURE)
	if measure != nil {
		log.Printf("Node (%s) InitMeasure: Measure client OK (already here)", n.nid)
	}
	stress = n.getOneClient(pb.ClientMode_STRESS)
	if stress != nil {
		log.Printf("Node (%s) InitMeasure: Stress client OK (already here)", n.nid)
	}

	for measure == nil || stress == nil {
		client := <- n.imClient
		if measure == nil && client.mode == pb.ClientMode_MEASURE {
			measure = client
			log.Printf("Node (%s) InitMeasure: Measure client OK", n.nid)
		}
		if stress == nil && client.mode == pb.ClientMode_STRESS {
			stress = client
			log.Printf("Node (%s) InitMeasure: Stress client OK", n.nid)
		}
	}


	n.changeIMState(1, "SETUP ZEROLOAD")
	stress.sendControlLoadGen(0)
	time.Sleep(2000 * time.Millisecond)
	measure.setPreMeasure(true)


	n.changeIMState(2, "MEASURE ZEROLOAD")
	zeroData := make([][]*pb.MeasureData, 5)
	for i := 0; i < 5; i++ {
		zeroData[i] = <- n.imMeasure
	}


	n.changeIMState(3, "SETUP FULLLOAD")
	measure.Stop()
	stress.sendControlLoadGen(3)
	time.Sleep(2000 * time.Millisecond)
	measure.setPreMeasure(true)


	n.changeIMState(4, "MEASURE FULLLOAD")
	fullData := make([][]*pb.MeasureData, 5)
	for i := 0; i < 5; i++ {
		fullData[i] = <- n.imMeasure
	}


	n.changeIMState(5, "DONE")
	stress.sendControlLoadGen(0)
	measure.setPreMeasure(false)


	// CALC SCORES
	n.score = &NodeScore{
		CpuZero: calcAvg(zeroData, "sng_cpu"),
		VmZero: calcAvg(zeroData, "sng_vm"),
		TimerZero: calcAvg(zeroData, "sng_timer"),
		CpuFull: calcAvg(fullData, "sng_cpu"),
		VmFull: calcAvg(fullData, "sng_vm"),
		TimerFull: calcAvg(fullData, "sng_timer"),
	}
	log.Printf("Node (%s) InitMeasure CPU score: %d -> %d", n.nid, n.score.CpuZero, n.score.CpuFull)
	log.Printf("Node (%s) InitMeasure VM score: %d -> %d", n.nid, n.score.VmZero, n.score.VmFull)
	log.Printf("Node (%s) InitMeasure TIMER score: %d -> %d", n.nid, n.score.TimerZero, n.score.TimerFull)

	// BROADCAST NEW CALCULATED SCORES
	n.app.browserGw.BroadcastNodeNewScoreMessage(n)
}

func (n *Node) NewMeasureData(data []*pb.MeasureData) bool {
	if n.imState < 2 {
		return true
	} else if n.imState == 2 {
		n.imMeasure <- data
		return true
	} else if n.imState == 3 {
		return true
	} else if n.imState == 4 {
		n.imMeasure <- data
		return true
	} else {
		return false
	}
}

func (n *Node) getClients(mode pb.ClientMode) []*Client {
	var res []*Client
	for _, client := range n.app.clients {
		if client.node == n && (mode == -1 || client.mode == mode) {
			res = append(res, client)
		}
	}
	return res
}

func (n *Node) getOneClient(mode pb.ClientMode) *Client {
	for _, client := range n.app.clients {
		if client.node == n && (mode == -1 || client.mode == mode) {
			return client
		}
	}
	return nil
}


func calcAvg(data [][]*pb.MeasureData, name string) uint64 {
	var sum uint64
	var count uint64

	sum = 0
	count = 0
	for _, dataLine := range data {
		for _, m := range dataLine {
			if m.Name == name {
				sum += m.Value
				count += 1
			}
		}
	}

	if count > 0 {
		return sum / count
	} else {
		return 0
	}
}