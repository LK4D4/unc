package main

import (
	"os"
	"os/exec"
	"syscall"
)

type Container struct {
	Args []string
	Uid  int
	Gid  int
}

func (c *Container) Start() error {
	cmd := exec.Command(c.Args[0], c.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      c.Uid,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      c.Gid,
				Size:        1,
			},
		},
	}
	return cmd.Run()
}
