package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/Sirupsen/logrus"
)

type Container struct {
	Args []string
	Uid  int
	Gid  int
}

func (c *Container) Start() error {
	cmd := &exec.Cmd{
		Path: os.Args[0],
		Args: append([]string{"unc-fork"}, c.Args...),
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUTS,
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

type Cfg struct {
	Path     string
	Args     []string
	Hostname string
}

var defaultCfg = Cfg{
	Hostname: "unc",
}

func setup(cfg Cfg) error {
	if err := syscall.Sethostname([]byte(cfg.Hostname)); err != nil {
		return fmt.Errorf("Sethostname: %v", err)
	}
	return nil
}

func execProc(cfg Cfg) error {
	logrus.Debugf("Execute %s", append([]string{cfg.Path}, cfg.Args[1:]...))
	return syscall.Exec(cfg.Path, cfg.Args, os.Environ())
}

func fork() error {
	logrus.Debug("Start fork")
	name, err := exec.LookPath(os.Args[1])
	if err != nil {
		return fmt.Errorf("LookPath: %v", err)
	}
	defaultCfg.Path = name
	defaultCfg.Args = os.Args[1:]
	if err := setup(defaultCfg); err != nil {
		return err
	}
	return execProc(defaultCfg)
}
