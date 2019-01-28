package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const suidNet = "unet"

func waitForIface() (netlink.Link, error) {
	logrus.Debugf("Starting to wait for network interface")
	start := time.Now()
	for {
		fmt.Printf(".")
		if time.Since(start) > 5*time.Second {
			fmt.Printf("\n")
			return nil, fmt.Errorf("failed to find veth interface in 5 seconds")
		}
		// get list of all interfaces
		lst, err := netlink.LinkList()
		if err != nil {
			fmt.Printf("\n")
			return nil, err
		}
		for _, l := range lst {
			// if we found "veth" interface - it's time to continue setup
			if l.Type() == "veth" {
				fmt.Printf("\n")
				return l, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func putIface(pid int) error {
	logrus.Debugf("Putting veth interface into container")
	cmd := exec.Command(suidNet, strconv.Itoa(pid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unet: err: %v", err)
	}
	logrus.Debugf("Network setup done")
	return nil
}

func setupIface(link netlink.Link, cfg Cfg) error {
	// up loopback
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("lo interface: %v", err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("up veth: %v", err)
	}
	addr, err := netlink.ParseAddr(cfg.IP)
	if err != nil {
		return fmt.Errorf("parse IP: %v", err)
	}
	return netlink.AddrAdd(link, addr)
}
