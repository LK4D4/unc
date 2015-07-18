package main

import (
	"log"
	"os"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	if os.Args[0] == "unc-fork" {
		if err := fork(); err != nil {
			log.Fatalf("Fork error: %v", err)
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
		log.Fatalf("Container start failed: %v", err)
	}
}
