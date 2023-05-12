package main

import (
	"fmt"
	"net"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

type TUN struct {
	water.Interface
}

func NewTUN(tunName string, mtu int, localAddr string, remoteAddr string) (*TUN, error) {
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: tunName,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error creating network device: %s", err)
	}

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, fmt.Errorf("error getting interface: %s", err)
	}

	_, localNetIP, err := net.ParseCIDR(localAddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing local address: %s", err)
	}

	_, remoteNetIP, err := net.ParseCIDR(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing remote address: %s", err)
	}

	addr := netlink.Addr{
		IPNet: localNetIP,
		Peer:  remoteNetIP,
	}

	err = netlink.AddrAdd(link, &addr)
	if err != nil {
		return nil, fmt.Errorf("error adding address: %s", err)
	}

	err = netlink.LinkSetMTU(link, mtu)
	if err != nil {
		return nil, fmt.Errorf("error setting link mtu: %s", err)
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return nil, fmt.Errorf("error setting link up: %s", err)
	}

	return &TUN{Interface: *iface}, nil
}
