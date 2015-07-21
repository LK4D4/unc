package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
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
