package zk

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/mlowicki/rhythm/conf"
	"github.com/mlowicki/rhythm/zkutil"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
)

const electionDir = "election"

type Coordinator struct {
	dir       string
	conn      *zk.Conn
	acl       func(perms int32) []zk.ACL
	ticket    string
	eventChan <-chan zk.Event
	cancel    context.CancelFunc
	sync.Mutex
}

func (coord *Coordinator) WaitUntilLeader() (context.Context, error) {
	isLeader, ch, err := coord.isLeader()
	if err != nil {
		return nil, err
	}
	if !isLeader {
		for {
			log.Println("Not elected as leader. Waiting...")
			<-ch
			isLeader, ch, err = coord.isLeader()
			if err != nil {
				return nil, err
			} else if isLeader {
				break
			}
		}
	}
	log.Println("Elected as leader")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	coord.Lock()
	coord.cancel = cancel
	coord.Unlock()
	return ctx, nil
}

func (coord *Coordinator) register() error {
	name, err := coord.conn.Create(coord.dir+"/"+electionDir+"/", []byte{}, zk.FlagEphemeral|zk.FlagSequence, coord.acl(zk.PermAll))
	if err != nil {
		return err
	}
	parts := strings.Split(name, "/")
	coord.Lock()
	coord.ticket = parts[len(parts)-1]
	coord.Unlock()
	return nil
}

func (coord *Coordinator) isLeader() (bool, <-chan zk.Event, error) {
	coord.Lock()
	ticket := coord.ticket
	coord.Unlock()
	if ticket == "" {
		err := coord.register()
		if err != nil {
			return false, nil, fmt.Errorf("Registration failed: %s", err)
		}
	}
	tickets, _, eventChan, err := coord.conn.ChildrenW(coord.dir + "/" + electionDir)
	if err != nil {
		return false, nil, fmt.Errorf("Failed getting registration tickets: %s", err)
	}
	coord.Lock()
	ticket = coord.ticket
	coord.Unlock()
	isLeader := false
	sort.Strings(tickets)
	if len(tickets) > 0 {
		if tickets[0] == ticket {
			isLeader = true
		}
	}
	log.Debugf("All registration tickets: %v", tickets)
	log.Debugf("My registration ticket: %s", ticket)
	for _, cur := range tickets {
		if ticket == cur {
			return isLeader, eventChan, nil
		}
	}
	return false, nil, fmt.Errorf("Registration ticket doesn't exist")
}

func (coord *Coordinator) initZK() error {
	exists, _, err := coord.conn.Exists(coord.dir)
	if err != nil {
		return err
	}
	if !exists {
		_, err = coord.conn.Create(coord.dir, []byte{}, 0, coord.acl(zk.PermAll))
		if err != nil {
			return err
		}
	}
	path := coord.dir + "/" + electionDir
	exists, _, err = coord.conn.Exists(path)
	if err != nil {
		return fmt.Errorf("Failed checking if election directory exists: %s", err)
	}
	if !exists {
		_, err = coord.conn.Create(path, []byte{}, 0, coord.acl(zk.PermAll))
		if err != nil {
			return fmt.Errorf("Failed creating election directory: %s", err)
		}
	}
	return nil
}

func New(c *conf.CoordinatorZK) (*Coordinator, error) {
	conn, eventChan, err := zk.Connect(c.Addrs, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed connecting to ZooKeeper: %s", err)
	}
	acl, err := zkutil.AddAuth(conn, &c.Auth)
	if err != nil {
		return nil, err
	}
	coord := Coordinator{
		conn:      conn,
		acl:       acl,
		dir:       "/" + c.Dir,
		eventChan: eventChan,
	}
	err = coord.initZK()
	if err != nil {
		conn.Close()
		return nil, err
	}
	go func() {
		for {
			select {
			case ev := <-coord.eventChan:
				log.Printf("ZooKeeper event: %s", ev)
				if ev.State == zk.StateDisconnected {
					log.Printf("Disconnected from ZooKeeper: %s", ev)
					coord.Lock()
					if coord.cancel != nil {
						coord.cancel()
						coord.cancel = nil
					}
					coord.Unlock()
				} else if ev.State == zk.StateExpired {
					log.Printf("Session expired: %s", ev)
					coord.Lock()
					coord.ticket = ""
					coord.Unlock()
				}
			}
		}
	}()
	return &coord, nil
}
