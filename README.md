# posse - Packets Over Shared StoragE

This is an experimenal tool that allows IP communication between two hosts through a shared storage device, which could be a FC or iSCSI LUN mapped to both hosts, or a virtual disk that is attached to two VMs, such as a multi-writer shared disk in VMware vSphere or an EBS Multi-Attach disk in AWS.

It is uncertain what are the the optimal use cases for this tool at present. However, it is possible that it could serve as an additional out-of-band link in high availability (HA) environments to better manage network isolation incidents, or circumvent restrictive firewalls.

It is anticipated that there may be instances of packet drops and other minor interruptions, although low-traffic applications such as SSH are currently functioning properly.

## Working Principle

This tool creates a TUN interface and writes incoming packets to a designated disk at a specified block number (`wblk`). Simultaneously, it polls the other designated block (`rblk`) for changes and writes them to the TUN interface.

## Example
```
host1# posse -disk /dev/sdb -rblk 0 -wblk 1 -addr 10.10.10.1/32 -peer 10.10.10.2/32
host2# posse -disk /dev/sdb -rblk 1 -wblk 0 -addr 10.10.10.2/32 -peer 10.10.10.1/32
host1# ping 10.10.10.2
ping -i 0.02 10.10.10.2
PING 10.10.10.2 (10.10.10.2) 56(84) bytes of data.
64 bytes from 10.10.10.2: icmp_seq=4 ttl=64 time=44.4 ms
64 bytes from 10.10.10.2: icmp_seq=9 ttl=64 time=29.4 ms
64 bytes from 10.10.10.2: icmp_seq=13 ttl=64 time=38.5 ms
^C
```

## Options

- `disk` - Path to disk used for sending and receiving packets. It is important to exercise caution when using this disk, as any data on it may be overwritten. Required.
- `tun` - Tunnel device name. Optional.
- `addr` - Local IP address for tunnel. Must be equal to `peer` value on the remote host. Required.
- `peer` - Remote IP address for tunnel. Must be equal to `addr` value on the remote host. Required.
- `rblk` - Disk block number to read packets from. Must be equal to `wblk` value on the remote host. Required.
- `wblk` - Disk block bumber to write packets to. Must be equal to `rblk` value on the remote host. Required.
- `txqlen` - Transmit queue length. Optional.
- `rxqlen` - Transmit queue length. Optional.
- `hz` - Frequency in Hz at which the writing and polling operations are performed. Must be equal to `hz` value on the remote host. Optional.