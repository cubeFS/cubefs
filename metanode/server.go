// Copyright 2018 The CubeFS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package metanode

import (
	"fmt"
	"io"
	"net"

	"github.com/xtaci/smux"

	"github.com/cubefs/cubefs/proto"
	"github.com/cubefs/cubefs/util"
	"github.com/cubefs/cubefs/util/log"
)

// StartTcpService binds and listens to the specified port.
func (m *MetaNode) startServer() (err error) {
	// initialize and start the server.
	m.httpStopC = make(chan uint8)

	addr := fmt.Sprintf(":%s", m.listen)
	if m.bindIp {
		addr = fmt.Sprintf("%s:%s", m.localAddr, m.listen)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go func(stopC chan uint8) {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			select {
			case <-stopC:
				return
			default:
			}
			if err != nil {
				continue
			}
			go m.serveConn(conn, stopC)
		}
	}(m.httpStopC)
	log.Info("start server over...")
	return
}

func (m *MetaNode) stopServer() {
	if m.httpStopC != nil {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("action[StopTcpServer] err:%v", r)
			}
		}()
		close(m.httpStopC)
	}
}

// Read data from the specified tcp connection until the connection is closed by the remote or the tcp service is down.
func (m *MetaNode) serveConn(conn net.Conn, stopC chan uint8) {
	defer func() {
		conn.Close()
		m.RemoveConnection()
	}()
	m.AddConnection()
	c := conn.(*net.TCPConn)
	c.SetKeepAlive(true)
	c.SetNoDelay(true)
	remoteAddr := conn.RemoteAddr().String()
	for {
		select {
		case <-stopC:
			return
		default:
		}
		p := &Packet{}
		if err := p.ReadFromConnWithVer(conn, proto.NoReadDeadlineTime); err != nil {
			if err != io.EOF {
				p.Span().Error("serve MetaNode: ", err.Error())
			}
			return
		}
		span := p.Span()
		if err := m.handlePacket(conn, p, remoteAddr); err != nil {
			span.Errorf("serve handlePacket fail: %v", err)
		}
		span.Info("tracks:", span.TrackLog())
		span.Finish()
	}
}

func (m *MetaNode) handlePacket(conn net.Conn, p *Packet, remoteAddr string) (err error) {
	// Handle request
	err = m.metadataManager.HandleMetadataOperation(conn, p, remoteAddr)
	return
}

func (m *MetaNode) startSmuxServer() (err error) {
	// initialize and start the server.
	m.smuxStopC = make(chan uint8)

	ipPort := fmt.Sprintf(":%s", m.listen)
	if m.bindIp {
		ipPort = fmt.Sprintf("%s:%s", m.localAddr, m.listen)
	}
	addr := util.ShiftAddrPort(ipPort, smuxPortShift)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go func(stopC chan uint8) {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			select {
			case <-stopC:
				return
			default:
			}
			if err != nil {
				continue
			}
			go m.serveSmuxConn(conn, stopC)
		}
	}(m.smuxStopC)
	log.Info("start smux server over ...")
	return
}

func (m *MetaNode) stopSmuxServer() {
	if smuxPool != nil {
		smuxPool.Close()
		log.Debug("action[stopSmuxServer] stop smux conn pool")
	}

	if m.smuxStopC != nil {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("action[stopSmuxServer],err:%v", r)
			}
		}()
		close(m.smuxStopC)
	}
}

func (m *MetaNode) serveSmuxConn(conn net.Conn, stopC chan uint8) {
	defer func() {
		conn.Close()
		m.RemoveConnection()
	}()
	m.AddConnection()
	c := conn.(*net.TCPConn)
	c.SetKeepAlive(true)
	c.SetNoDelay(true)
	remoteAddr := conn.RemoteAddr().String()

	var sess *smux.Session
	var err error
	sess, err = smux.Server(conn, smuxPoolCfg.Config)
	if err != nil {
		log.Errorf("action[serveSmuxConn] failed to serve smux connection, err(%v)", err)
		return
	}
	defer sess.Close()

	for {
		select {
		case <-stopC:
			return
		default:
		}

		stream, err := sess.AcceptStream()
		if err != nil {
			if util.FilterSmuxAcceptError(err) != nil {
				log.Errorf("action[startSmuxService] failed to accept, err: %s", err)
			} else {
				log.Errorf("action[startSmuxService] accept done, err: %s", err)
			}
			break
		}
		go m.serveSmuxStream(stream, remoteAddr, stopC)
	}
}

func (m *MetaNode) serveSmuxStream(stream *smux.Stream, remoteAddr string, stopC chan uint8) {
	for {
		select {
		case <-stopC:
			return
		default:
		}

		p := &Packet{}
		if err := p.ReadFromConnWithVer(stream, proto.NoReadDeadlineTime); err != nil {
			if err != io.EOF {
				p.Span().Error("serve MetaNode: ", err.Error())
			}
			return
		}
		span := p.Span()
		if err := m.handlePacket(stream, p, remoteAddr); err != nil {
			span.Errorf("serve handlePacket fail: %v", err)
		}
		span.Debug("tracks:", span.TrackLog())
		span.Finish()
	}
}
