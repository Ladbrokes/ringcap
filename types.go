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
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"io"
)

type DataPacket struct {
	CaptureInfo gopacket.CaptureInfo
	Data        []byte
}

type PacketRing struct {
	inputChannel  <-chan gopacket.Packet
	outputChannel chan DataPacket
	ignore bool
}

func NewPacketRing(handle gopacket.PacketDataSource, linkType layers.LinkType, packetLimit int) *PacketRing {
	packetSource := gopacket.NewPacketSource(handle, linkType)
	inputChannel := packetSource.Packets()

	outputChannel := make(chan DataPacket, packetLimit)
	return &PacketRing{inputChannel, outputChannel, false}
}

func (r *PacketRing) Run() {
	for v := range r.inputChannel {
		dp := DataPacket{v.Metadata().CaptureInfo, v.Data()}
		if !r.ignore {
			select {
			case r.outputChannel <- dp:
			default:
				select {
				case <-r.outputChannel:
				default:
				}
				r.outputChannel <- dp
			}
		}
	}
	close(r.outputChannel)
}

func (r *PacketRing) Count() int {
	return len(r.outputChannel)
}

func (r *PacketRing) WritePackets(writer io.Writer, snapLen int, linkType layers.LinkType) error {
	r.ignore = true;
	defer func(r *PacketRing) {
		r.ignore = false
	}(r);

	pwriter := pcapgo.NewWriter(writer)
	pwriter.WriteFileHeader(uint32(snapLen), linkType)

	for {
		select {
		case data := <-r.outputChannel:
			err := pwriter.WritePacket(data.CaptureInfo, data.Data)
			if err != nil {
				return err
			}
		default:
			return nil
		}
	}
}
