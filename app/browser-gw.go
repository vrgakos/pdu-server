package app

import (
	"gopkg.in/olahol/melody.v1"
	"github.com/labstack/gommon/log"
	"encoding/json"
	pb "pdu-server/protos"
)

type BrowserGw struct {
	app *App
    mr *melody.Melody
}

func NewBrowserGw(app *App) *BrowserGw {
	bgw := &BrowserGw{
		app: app,
		mr: melody.New(),
	}

	bgw.mr.HandleConnect(func(s *melody.Session) {
		//s.Request.URL.Query().Get("id")
		log.Info("BGW", "somebody connected")

		// TODO: CONCURRENT ERROR MAY OCCURS
		for _, n := range bgw.app.nodes {
			s.Write(getNodeConnectedMessage(n))
		}
		for _, c := range bgw.app.clients {
			s.Write(getClientConnectedMessage(c))
			if c.pod != nil {
				s.Write(getLoadChangedMessage(c.pod.id, c.pod.loadGenLevel))
			}
		}

		s.Write(getLoadedMessage())
	})

	bgw.mr.HandleDisconnect(func(s *melody.Session) {
		log.Info("BGW", "somebody disconnected")
	})

	bgw.mr.HandleMessage(func(s *melody.Session, msg []byte) {
//		log.Info("BGW", "msg", msg)

		type message struct {
			T string				`json:"t"`
			D *json.RawMessage		`json:"d"`
		}

		var m message
		json.Unmarshal(msg, &m)

		switch m.T {
		case "loadGen":
			type data struct {
				Id string			`json:"id"`
				Value int			`json:"value"`
			}

			var d data
			json.Unmarshal(*m.D, &d)

			bgw.app.getPod(d.Id).setLoadGenLevel(d.Value)
			bgw.mr.BroadcastOthers(getLoadChangedMessage(d.Id, d.Value), s)

		case "initMeasure":
			type data struct {
				Nid string			`json:"nid"`
			}

			var d data
			json.Unmarshal(*m.D, &d)

			go bgw.app.getNode(d.Nid).initMeasure()
		}

	})


	return bgw
}


func (bgw *BrowserGw) GetMelodyRouter() *melody.Melody {
	return bgw.mr
}

func (bgw *BrowserGw) BroadcastClientConnected(client *Client) {
	bgw.mr.Broadcast(getClientConnectedMessage(client))
}

func (bgw *BrowserGw) BroadcastClientDisconnected(client *Client) {
	bgw.mr.Broadcast(getClientDisconnectedMessage(client))
}

func (bgw *BrowserGw) BroadcastClientMeasureData(client *Client, collect *pb.CollectRequest) {
	bgw.mr.Broadcast(getClientMeasureDataMessage(client, collect))
}

func (bgw *BrowserGw) BroadcastNodeMeasureData(node *Node, collect *pb.NodeCollectRequest) {
	bgw.mr.Broadcast(getNodeMeasureDataMessage(node, collect))
}

func (bgw *BrowserGw) BroadcastNodeConnected(node *Node) {
	bgw.mr.Broadcast(getNodeConnectedMessage(node))
}

func (bgw *BrowserGw) BroadcastNodeDisconnected(node *Node) {
	bgw.mr.Broadcast(getNodeDisconnectedMessage(node))
}

func (bgw *BrowserGw) BroadcastNodeNewScoreMessage(node *Node) {
	bgw.mr.Broadcast(getNodeNewScoreMessage(node))
}

func (bgw *BrowserGw) BroadcastNodeIMMessage(node *Node) {
	bgw.mr.Broadcast(getNodeIMMessage(node))
}

type message struct {
	T string		`json:"t"`
	D interface{}	`json:"d"`
}


func getLoadChangedMessage(id string, value int) []byte {
	type data struct {
		Id string					`json:"id"`
		Value int					`json:"value"`
	}

	msg := message{
		T: "lc",
		D: data{
			Id: id,
			Value: value,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getClientConnectedMessage(client *Client) []byte {
	type data struct {
		Id string	`json:"id"`
		Cid uint32	`json:"cid"`
		Nid string 	`json:"nid"`
		Name string	`json:"name"`
		Mode int32  `json:"mode"`
	}

	msg := message{
		T: "cc",
		D: data{
			Id: client.pod.id,
			Cid: client.cid,
			Nid: client.node.nid,
			Name: client.name,
			Mode: int32(client.mode),
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getClientDisconnectedMessage(client *Client) []byte {
	type data struct {
		Id string	`json:"id"`
		Cid uint32	`json:"cid"`
		Name string	`json:"name"`
		Mode int32  `json:"mode"`
	}

	msg := message{
		T: "cd",
		D: data{
			Id: client.pod.id,
			Cid: client.cid,
			Name: client.name,
			Mode: int32(client.mode),
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getLoadedMessage() []byte {
	msg := message{
		T: "loaded",
		D: nil,
	}

	m, _ := json.Marshal(msg)
	return m
}

func getClientMeasureDataMessage(client *Client, collect *pb.CollectRequest) []byte {
	type data struct {
		Id string					`json:"id"`
		Cid uint32					`json:"cid"`
		Time int64					`json:"time"`
		Measure []*pb.MeasureData	`json:"measure"`
	}

	msg := message{
		T: "m",
		D: data{
			Id: client.pod.id,
			Cid: client.cid,
			Time: collect.Time,
			Measure: collect.Data,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getNodeMeasureDataMessage(node *Node, collect *pb.NodeCollectRequest) []byte {
	type data struct {
		Nid string					`json:"nid"`
		Time int64					`json:"time"`
		Measure []*pb.MeasureData	`json:"measure"`
	}

	msg := message{
		T: "nm",
		D: data{
			Nid: node.nid,
			Time: collect.Time,
			Measure: collect.Data,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getNodeNewScoreMessage(node *Node) []byte {
	type data struct {
		Nid string			`json:"nid"`
		Score *NodeScore	`json:"score"`
	}

	msg := message{
		T: "ns",
		D: data{
			Nid: node.nid,
			Score: node.score,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getNodeIMMessage(node *Node) []byte {
	type data struct {
		Nid string			`json:"nid"`
		State int			`json:"state"`
	}

	msg := message{
		T: "ni",
		D: data{
			Nid: node.nid,
			State: node.imState,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}


func getNodeConnectedMessage(node *Node) []byte {
	type data struct {
		Nid string			`json:"nid"`
		Name string			`json:"name"`
		Score *NodeScore	`json:"score"`
		State int			`json:"state"`
	}

	msg := message{
		T: "nc",
		D: data{
			Nid: node.nid,
			Name: node.name,
			Score: node.score,
			State: node.imState,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}

func getNodeDisconnectedMessage(node *Node) []byte {
	type data struct {
		Nid string	`json:"nid"`
		Name string	`json:"name"`
	}

	msg := message{
		T: "nd",
		D: data{
			Nid: node.nid,
			Name: node.name,
		},
	}

	m, _ := json.Marshal(msg)
	return m
}
