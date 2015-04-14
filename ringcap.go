/*
 *   Ringcap - pcap ringbuffer
 *   Copyright (c) 2015 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd.
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *   Author: Shannon Wynter <http://fremnet.net/contact>
 */
package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/layers"
	"github.com/op/go-logging"
	"github.com/pivotal-golang/bytefmt"
	"io"
	"net"
	"os"
	"path"
	"time"
)

var (
	bindAddr    = "127.0.0.1:4231"
	dumpHost    = "127.0.0.1:4231"
	packetLimit = 10000
	snapLen     = 65535
	iface       = "eth0"
	filter      = ""

	listener = false
	savePath = ""

	printVersion = false
)

var log = logging.MustGetLogger("ringcap")

func init() {
	savePath, _ = os.Getwd()
	flag.StringVar(&bindAddr, "bind-addr", bindAddr, "Address to bind to")
	flag.StringVar(&dumpHost, "dump-host", dumpHost, "Host Address to dump to")
	flag.IntVar(&packetLimit, "packet-limit", packetLimit, "Maximum packet count")
	flag.IntVar(&snapLen, "snaplen", snapLen, "Maximum packet size in bytes")
	flag.StringVar(&iface, "interface", iface, "interface to monitor")
	flag.StringVar(&filter, "filter", filter, "Optional filter")
	flag.BoolVar(&printVersion, "version", printVersion, "Display the current version")
	flag.BoolVar(&listener, "listen", listener, "Listen for dumps")
	flag.StringVar(&savePath, "save-path", savePath, "Where to store dumps (when listening)")
}

func initTrigger() (chan bool, error) {
	ch := make(chan bool, 1)

	addr, err := net.ResolveUDPAddr("udp4", bindAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	b := make([]byte, UDP_PACKET_SIZE)

	go func() {
		for {
			_, _, err := conn.ReadFromUDP(b)
			if err != nil {
				log.Warning("%s", err)
				continue
			}
			ch <- true
		}
	}()

	return ch, nil
}

func runDumper() {
	consoleBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackend("")
	if err != nil {
		log.Fatal(err)
	}

	logging.SetBackend(consoleBackend, syslogBackend)
	logging.SetFormatter(logging.MustStringFormatter("%{color}%{time:15:04:05.000000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"))

	log := logging.MustGetLogger("main")

	maxMemory := bytefmt.ByteSize(uint64(packetLimit * snapLen))
	log.Info("Preparing to capture on %s with a buffer of %d and snaplen of %d - Maximum memory usage %s", iface, packetLimit, snapLen, maxMemory)
	log.Info("On any udp packet sent to %s pcap formatted data will be sent to tcp://%s", bindAddr, dumpHost)

	triggerChan, err := initTrigger()
	if err != nil {
		log.Fatal(err)
	}

	handle, err := pcap.OpenLive(iface, int32(snapLen), false, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	if filter != "" {
		log.Info("Using filter \"%s\"", filter)
		if err := handle.SetBPFFilter(filter); err != nil { // optional
			log.Fatal("Unable to compile and set filter: ", err)
		}
	}

	pr := NewPacketRing(handle, handle.LinkType(), packetLimit)
	go pr.Run()

	for {
		select {
		case <-triggerChan:
			sendDump(pr, handle.LinkType())
		}
	}

}

func sendDump(pr *PacketRing, linkType layers.LinkType) {
	log.Info("Dumping %d captured packets to %s", pr.Count(), dumpHost)
	conn, err := net.Dial("tcp", dumpHost)
	if err != nil {
		log.Error("Unable to connect to %s: %s", dumpHost, err)
		return
	}
	defer conn.Close()

	err = pr.WritePackets(conn, snapLen, linkType)
	if err != nil {
		log.Error("Failed to dump to %s: %s", dumpHost, err)
		return
	}
}

func handleDump(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().(*net.TCPAddr)
	remoteIP := remote.IP.String()
	port := remote.Port
	log.Info("Connection from %s:%d", remoteIP, port)

	filename := path.Join(savePath, fmt.Sprintf("%s-%d-%d.pcap", remoteIP, time.Now().Unix(), port))

	file, err := os.Create(filename)
	if err != nil {
		log.Error("Unable to create ", filename, ": ", err)
		return
	}
	defer file.Close()

	bytes, err := io.Copy(file, conn)
	if err != nil {
		log.Error("Problem saving dump to ", filename, ": ", err)
		return
	}
	log.Info("Dump \"%s\" created - %s", path.Base(filename), bytefmt.ByteSize(uint64(bytes)))
}

func runListener() {
	l, err := net.Listen("tcp4", bindAddr)
	if err != nil {
		log.Fatal("Error listening: ", err)
	}
	defer l.Close()

	log.Info("Listening on %s for dumps", bindAddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error accepting: ", err)
			continue
		}
		go handleDump(conn)
	}
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Printf("ringcap %s\n", VERSION)
		os.Exit(0)
	}

	if listener {
		runListener()
	} else {
		runDumper()
	}
}
