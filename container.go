package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/Sirupsen/logrus"
)

const ipTmpl = "10.100.42.%d/24"

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
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
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
	if err := cmd.Start(); err != nil {
		return err
	}
	logrus.Debugf("container PID: %d", cmd.Process.Pid)
	if err := putIface(cmd.Process.Pid); err != nil {
		return err
	}
	return cmd.Wait()
}

type Mount struct {
	Source string
	Target string
	Fs     string
	Flags  int
	Data   string
}

type Cfg struct {
	Path     string
	Args     []string
	Hostname string
	Mounts   []Mount
	Rootfs   string
	IP       string
}

var defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

var defaultCfg = Cfg{
	Hostname: "unc",
	Mounts: []Mount{
		{
			Source: "proc",
			Target: "/proc",
			Fs:     "proc",
			Flags:  defaultMountFlags,
		},
		{
			Source: "tmpfs",
			Target: "/dev",
			Fs:     "tmpfs",
			Flags:  syscall.MS_NOSUID | syscall.MS_STRICTATIME,
			Data:   "mode=755",
		},
	},
	Rootfs: "/home/moroz/project/busybox",
}

func pivotRoot(root string) error {
	// we need this to satisfy restriction:
	// "new_root and put_old must not be on the same filesystem as the current root"
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	// create rootfs/.pivot_root as path for old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	logrus.Debugf("Pivot root dir: %s", pivotDir)
	logrus.Debugf("Pivot root to %s", root)
	// pivot_root to rootfs, now old_root is mounted in rootfs/.pivot_root
	// mounts from it still can be seen in `mount`
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// change working directory to /
	// it is recommendation from man-page
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	// path to pivot root now changed, update
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root(which is now /.pivot_root) with all submounts
	// now we have only mounts that we mounted ourself in `mount`
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// remove temporary directory
	return os.Remove(pivotDir)
}

func mount(cfg Cfg) error {
	for _, m := range cfg.Mounts {
		target := filepath.Join(cfg.Rootfs, m.Target)
		logrus.Debugf("Mount %s to %s", m.Source, target)
		if err := syscall.Mount(m.Source, target, m.Fs, uintptr(m.Flags), m.Data); err != nil {
			return fmt.Errorf("failed to mount %s to %s: %v", m.Source, target, err)
		}
	}
	return nil
}

func setup(cfg Cfg) error {
	if err := mount(cfg); err != nil {
		return err
	}
	if err := pivotRoot(cfg.Rootfs); err != nil {
		return fmt.Errorf("Pivot root error: %v", err)
	}
	if err := syscall.Sethostname([]byte(cfg.Hostname)); err != nil {
		return fmt.Errorf("Sethostname: %v", err)
	}
	return nil
}

func execProc(cfg Cfg) error {
	logrus.Debugf("Execute %s", append([]string{cfg.Path}, cfg.Args[1:]...))
	return syscall.Exec(cfg.Path, cfg.Args, os.Environ())
}

func fillCfg() error {
	name, err := exec.LookPath(os.Args[1])
	if err != nil {
		return fmt.Errorf("LookPath: %v", err)
	}
	defaultCfg.Path = name
	defaultCfg.Args = os.Args[1:]
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error get working dir: %v", err)
	}
	defaultCfg.Rootfs = wd
	// choose ip
	defaultCfg.IP = fmt.Sprintf(ipTmpl, rand.Intn(253)+2)
	return nil
}

func fork() error {
	logrus.Debug("Start fork")
	if err := fillCfg(); err != nil {
		return err
	}
	if err := setup(defaultCfg); err != nil {
		return err
	}
	lnk, err := waitForIface()
	if err != nil {
		return err
	}
	if err := setupIface(lnk, defaultCfg); err != nil {
		return err
	}
	return execProc(defaultCfg)
}
