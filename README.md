# ringcap

`ringcap` is a diagnostic tool that allows you to capture packets in a ringbuffer to be dumped to a secondary host on a UDP trigger.

A useful tool for tracking down seemingly random network/protocol issues.

## Arguments

### -listen

Pass this argument if this is the dump host to cause it to listen on tcp

### -bind-addr `IP:PORT`

The address to bind to (tcp or udp depending on --listen)

Both the `IP` and the `PORT` are required

### -dump-host `HOST:PORT`

The host to dump to (the one that is running --listen)

Both the `HOST` and the `PORT` are required

Ignored with -listen

### -interface `eth0`

The interface to inspect traffic on (eth0, eth1, bind0, all)

Ignored with -listen

### -packet-limit `10000`

The maximum number of packets to store in memory

Ignored with -listen

### -snaplen `1-65535`

Snarf snaplen bytes of data from each packet rather than the default of 65535 bytes.

Ignored with -listen

### -save-path `/path/to/save`

Where to save the files captured with -listen, defaults to current working directory

### -filter `"pcap filter string"`

Pass a filter to pcap (eg: `"not port 22"`)

Ignored with -listen

## Usage

Set up a listener on a central host (10.1.1.99) - on this host 4231 is a tcp socket
```
ringcap` -listen -bind-addr 0.0.0.0:4231 -save-path /tmp
```

Setup a capture ringbuffer on the troublesome hosts (10.1.1.10) - on this host 4231 is a udp socket
```
ringcap -bind-addr 0.0.0.0:4231 -dump-host 10.1.1.99:4231 -interface bind0
```

Send any udp packet to the ringbuffer to trigger a dump
```
echo -n "hi" | nc -u4 -w1 10.1.1.10 4231
```

## Dependancies

* [Libpcap](http://www.tcpdump.org/#latest-release) (libpcap-devel)
* [Godep](https://github.com/tools/godep)

## Building

```
mkdir -p "${GOPATH}/src/github.com/Ladbrokes"
cd "${GOPATH}/src/github.com/Ladbrokes"
git clone https://github.com/Ladbrokes/ringcap.git
cd ringcap
godep go build .
```

##License

Copyright (c) 2015 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. Licensed under GPL3. See the [LICENSE.md](LICENSE.md) file for a copy of the license.
