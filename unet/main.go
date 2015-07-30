// You should run this binary with suid set.
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	bridgeName = "unc0"
	vethPrefix = "uv"
	ipAddr     = "10.100.42.1/24"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func createBridge() error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation: %v", err)
	}
	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		return fmt.Errorf("parse address %s: %v", ipAddr, err)
	}
	if err := netlink.AddrAdd(br, addr); err != nil {
		return fmt.Errorf("add address %v to bridge: %v", addr, err)
	}
	if err := netlink.LinkSetUp(br); err != nil {
		return err
	}
	return nil
}

func createVethPair(pid int) error {
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	x1, x2 := rand.Intn(10000), rand.Intn(10000)
	parentName := fmt.Sprintf("%s%d", vethPrefix, x1)
	peerName := fmt.Sprintf("%s%d", vethPrefix, x2)
	la := netlink.NewLinkAttrs()
	la.Name = parentName
	la.MasterIndex = br.Attrs().Index
	vp := &netlink.Veth{LinkAttrs: la, PeerName: peerName}
	if err := netlink.LinkAdd(vp); err != nil {
		return fmt.Errorf("veth pair creation %s <-> %s: %v", parentName, peerName, err)
	}
	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("get peer interface: %v", err)
	}
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		return fmt.Errorf("move peer to ns of %d: %v", pid, err)
	}
	if err := netlink.LinkSetUp(vp); err != nil {
		return err
	}
	return nil
}

func main() {
	pid := 1
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		pid = p
	}
	if err := createBridge(); err != nil {
		log.Fatal(err)
	}
	if err := createVethPair(pid); err != nil {
		log.Fatal(err)
	}
}
