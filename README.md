# ringcap

`ringcap` is a diagnostic tool that allows you to capture packets in a ringbuffer for dumping on a trigger

A useful tool for tracking down seemingly random network/protocol issues.

Set up a listener on a central host
```
ringcap` -listen -bind-addr 0.0.0.0:4231 -save-path /tmp
```

Setup a capture ringbuffer on the troublesome hosts
```
ringcap -bind-addr 0.0.0.0:4231 -dump-host 10.1.1.99:4231 -interface bind0
```

Send any udp packet to the ringbuffer to trigger a dump
```
echo -n "hi" | nc -u4 -w1 10.1.1.10 4231
```

##License

Copyright (c) 2015 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. Licensed under GPL2. See the [LICENSE.md](LICENSE.md) file for a copy of the license.
