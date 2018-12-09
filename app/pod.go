package app

import (
	pb "pdu-server/protos"
	"sync"
	"github.com/labstack/gommon/log"
)

type Pod struct {
	id string
	lock sync.RWMutex
	app *App
	clients map[pb.ClientMode]*Client
	loadGenLevel int
}

func NewPod(app *App, id string) *Pod {
	return &Pod{
		id: id,
		app: app,
		clients: make(map[pb.ClientMode]*Client),
		loadGenLevel: 0,
	}
}

func (p *Pod) getClient(mode pb.ClientMode) *Client {
	p.lock.RLock()
	client, found := p.clients[mode]
	p.lock.RUnlock()

	if ! found {
		client = NewClient(p, mode)
		p.lock.Lock()
		p.clients[mode] = client
		p.lock.Unlock()
	}

	return client
}

func (p *Pod) removeClient(client *Client) {
	p.lock.Lock()
	delete(p.clients, client.mode)
	p.lock.Unlock()
}

func (p *Pod) setLoadGenLevel(level int) {
	p.loadGenLevel = level
	p.getClient(pb.ClientMode_STRESS).sendControlLoadGen(level)
	log.Printf("Change id=%s loadGenLevel to %d\n", p.id, level)
}