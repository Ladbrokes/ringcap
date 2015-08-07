Building a .deb package for Ubuntu Precise
==========================================

To build a new .deb package for a different version of Ubuntu, run up a new docker container in any Linux VM running under vagrant:

    $ docker pull ubuntu:precise
    $ docker run -t -i -v /vagrant:/vagrant ubuntu:precise
    root@dcc3e7bf2514:/# 

Inside the container, run the following commands:

    cd
    apt-get update
    apt-get install -y build-essential git libpcap-dev checkinstall
    wget https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz && tar xvf godeb-amd64.tar.gz
    ./godeb list
    ./godeb install 1.4.2
    export GOPATH=~/gocode
    go get github.com/tools/godep
    mkdir -p "${GOPATH}/src/github.com/Ladbrokes"
    cd "${GOPATH}/src/github.com/Ladbrokes"
    git clone https://github.com/Ladbrokes/ringcap.git
    cd ringcap
    ~/gocode/bin/godep go build .
    checkinstall -D cp ringcap /usr/sbin/ringcap

Answers for checkinstall script:

    End your description with an empty line or EOF.
    >> Ringcap is a ring buffer for libpcap
    >>
     
    Enter a number to change any of them or press ENTER to continue: 10
    Enter the additional requirements:
    >> libpcap0.8
     
    Enter a number to change any of them or press ENTER to continue: 0
    Enter the maintainer's name and e-mail address:
    >> your email foo@bar.com
     
    This package will be built according to these values:
    0 -  Maintainer: [ your email foo@bar.com ]
    1 -  Summary: [ Ringcap is a ring buffer for libpcap ]
    2 -  Name:    [ ringcap ]
    3 -  Version: [ 20150805 ]
    4 -  Release: [ 1 ]
    5 -  License: [ GPL ]
    6 -  Group:   [ checkinstall ]
    7 -  Architecture: [ amd64 ]
    8 -  Source location: [ ringcap ]
    9 -  Alternate source location: [  ]
    10 - Requires: [ libpcap0.8 ]
    11 - Provides: [ ringcap ]
    12 - Conflicts: [  ]
    13 - Replaces: [  ]

You should now have a debian package in the build directory which can be copied over to destination hosts and installed.

    dpkg -c ringcap_20150805-1_amd64.deb
    cp ringcap_20150805-1_amd64.deb /vagrant/
    dpkg -i ringcap_20150805-1_amd64.deb

