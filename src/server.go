
package gobtc

import (
	"net"
	"log"
	"os"
	"container/list"
)

type Server struct {
	waitPeerHandler chan bool
	waitListenerHandler chan bool
	incomingPeers chan *Peer
	quitingPeers chan *Peer
	listener net.Listener
	log *log.Logger
	maxPeers int
}

func New(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		make(chan bool),
		make(chan bool),
		make(chan *Peer),
		make(chan *Peer),
		listener,
		log.New(os.Stderr, "gobtc: ", log.LstdFlags),
		10,
	}
	return s, nil
}


func (s *Server) Start() {
	go s.peerHandler()
	go s.listenerHandler()
}


func (s *Server) SetLogger(log *log.Logger) {
	s.log = log
}

func (s *Server) Wait() {
	<- s.waitPeerHandler
	<- s.waitListenerHandler
}

func (s *Server) peerHandler() {
	var peers *list.List = list.New()

	for {
		select {
		case peer := <- s.incomingPeers:
			if (peers.Len() >= s.maxPeers) {
				peer.conn.Close();
			}

			peers.PushBack(peer)
			s.log.Printf("Added peer %s", peer.conn.RemoteAddr())
			go peer.handler()

		case peer := <- s.quitingPeers:
			// TODO: remove peer
			found := false
			for e := peers.Front(); e != nil; e = e.Next() {
				tpeer := e.Value.(*Peer);
				if tpeer == peer {
					peers.Remove(e);
					s.log.Printf("Removed peer %s", peer.conn.RemoteAddr())
					found = true
				}
			}
			if (!found) {
				s.log.Printf("assert error: quiting peer not found on the list")
			}
		}
	}
	s.waitPeerHandler <- true
}

func (s *Server) AddPeer(conn net.Conn) {
	peer := &Peer{
		s,
		conn,
		make(chan bool),
	}
	s.incomingPeers <- peer
}

func (s *Server) listenerHandler() {
	s.log.Printf("Listening on %s", s.listener.Addr())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		s.AddPeer(conn)
	}
	s.waitListenerHandler <- true
}
