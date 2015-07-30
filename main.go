package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if os.Args[0] == "unc-fork" {
		if err := fork(); err != nil {
			logrus.Fatalf("Fork error: %v", err)
		}
		os.Exit(0)
	}
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{os.Getenv("SHELL")}
	}
	c := &Container{
		Args: args,
		Uid:  os.Getuid(),
		Gid:  os.Getgid(),
	}
	if err := c.Start(); err != nil {
		logrus.Fatalf("Container start failed: %v", err)
	}
}
