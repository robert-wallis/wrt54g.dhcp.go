wrt54g.dhcp.go
==============

List the clients on your Linksys/Cisco WRT54G router with stock/updated firmware.

usage
=====

    $ ./clients -ip=10.0.0.1 -u=admin -p=password
	Querying 10.0.0.1 for the DHCP client list.
	Obsidian 10.0.0.81 00:01:02:03:04:05 22:57:03 81
	  Sharpy 10.0.0.82 00:01:02:03:04:06 18:21:54 82
	KuaiMing 10.0.0.83 00:01:02:03:04:03 16:43:05 83
	KuaiMing 10.0.0.84 00:01:02:03:04:01 16:43:12 84

building
========

    go build clients.go

