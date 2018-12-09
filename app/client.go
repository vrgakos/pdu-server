package app

import (
	pb "pdu-server/protos"
	"strings"
)

type Client struct {
	pod *Pod
	node *Node
	name string
	mode pb.ClientMode
	cid uint32

	preMeasure bool

	control chan *pb.ClientControlResponse
}

func NewClient(pod *Pod, mode pb.ClientMode) *Client {
	return &Client{
		preMeasure: false,
		pod: pod,
		mode: mode,
	}
}

func (c *Client) sendControlLoadGen(value int) {
	if c.mode != pb.ClientMode_STRESS {
		return
	}

	ccr := &pb.ClientControlResponse{
		Command: "stress-ng",
		Enabled: true,
		Repeat: false,
	}

	switch value {
	case 0:
		ccr.Enabled = false
	case 1:
		if strings.HasPrefix(c.name, "i") {
			ccr.Args = []string{"--cpu", "0", "-l", "30"}
		} else {
			ccr.Args = []string{"--cpu", "0", "-l", "10"}
		}
	case 2:
		ccr.Args = []string{"--pipe", "1", "--timer", "1"}
	case 3:
		ccr.Args = []string{"--cpu", "4", "--vm", "1", "--pipe", "1", "--timer", "1"}
	}

	c.control <- ccr
}

func (c *Client) setPreMeasure(pre bool) {
	c.preMeasure = pre
	c.sendControlMeasure()
}

func (c *Client) sendControlMeasure() {
	if c.mode != pb.ClientMode_MEASURE {
		return
	}

	var ccr *pb.ClientControlResponse

	if c.preMeasure {
		// PRE MEASURE
		ccr = &pb.ClientControlResponse{
			Command: "stress-ng",
			Args: []string{"--metrics", "--cpu", "1", "--vm", "1", "-T1", "--vm-bytes", "20m", "-t2"},
			Enabled: true,
			Repeat: true,
			RepeatDelay: 1000,
		}
	} else {
		// NORMAL MEASURE
		ccr = &pb.ClientControlResponse{
			Command: "stress-ng",
			Args: []string{"--metrics", "--cpu", "1", "--vm", "1", "-T1", "--vm-bytes", "20m", "-t2"},
			Enabled: true,
			Repeat: true,
			RepeatDelay: 2000,
		}
	}

	c.control <- ccr
}

func (c *Client) Stop() {
	c.control <- &pb.ClientControlResponse{
		Enabled: false,
	}
}