package app

import (
	"sync"
	"fmt"
)

type App struct {
	lock sync.RWMutex
	nodes map[string]*Node
	pods map[string]*Pod

	clientLock sync.RWMutex
	cidCounter uint32
	clients map[uint32]*Client

	browserGw *BrowserGw
}

func NewApp() *App {
	a := &App{
		nodes: make(map[string]*Node),
		pods: make(map[string]*Pod),
		cidCounter: 0,
		clients: make(map[uint32]*Client),
	}
	a.browserGw = NewBrowserGw(a)
	return a
}

func (a *App) GetBrowserGw() *BrowserGw {
	return a.browserGw
}

func (a *App) getNode(nid string) *Node {
	a.lock.RLock()
	node, found := a.nodes[nid]
	a.lock.RUnlock()

	if ! found {
		node = NewNode(a, nid)
		a.lock.Lock()
		a.nodes[nid] = node
		a.lock.Unlock()
	}

	return node
}

func (a *App) removeNode(node *Node) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.nodes, node.nid)
	return nil
}

func (a *App) getPod(id string) *Pod {
	a.lock.RLock()
	pod, found := a.pods[id]
	a.lock.RUnlock()

	if ! found {
		pod = NewPod(a, id)
		a.lock.Lock()
		a.pods[id] = pod
		a.lock.Unlock()
	}

	return pod
}


func (a *App) getClient(cid uint32) *Client {
	a.clientLock.RLock()
	defer a.clientLock.RUnlock()

	if client, found := a.clients[cid]; found {
		return client
	}

	return nil
}

func (a *App) assignCid(client *Client) error {
	if client.cid != 0 {
		return fmt.Errorf("Already assigned")
	}

	a.clientLock.Lock()
	defer a.clientLock.Unlock()

	a.cidCounter++
	client.cid = a.cidCounter
	a.clients[client.cid] = client
	return nil
}

func (a *App) deAssignCid(client *Client) error {
	if client.cid == 0 {
		return fmt.Errorf("Not assigned")
	}

	a.clientLock.Lock()
	defer a.clientLock.Unlock()

	delete(a.clients, client.cid)
	return nil
}


func (a *App) clientConnected(client *Client) {
	// ASSIGN CID
	a.assignCid(client)
	client.node.ClientConnected(client)

	// BROADCAST TO BROWSERS
	a.GetBrowserGw().BroadcastClientConnected(client)
}

func (a *App) clientDisconnected(client *Client) {
	// DEASSIGN CID
	a.deAssignCid(client)
	client.pod.removeClient(client)

	// BROADCAST TO BROWSERS
	a.GetBrowserGw().BroadcastClientDisconnected(client)
}

func (a *App) nodeConnected(node *Node) {
	// ADD TO MAP

	// BROADCAST TO BROWSERS
	a.GetBrowserGw().BroadcastNodeConnected(node)
}

func (a *App) nodeDisconnected(node *Node) {
	// REMOVE FROM MAP
	a.removeNode(node)

	// BROADCAST TO BROWSERS
	a.GetBrowserGw().BroadcastNodeDisconnected(node)
}





