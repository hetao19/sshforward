package main

import (
	"flag"
	"sshforward/g"
)

var (
	config = flag.String("c", "./cfg.json", "config json path")
)

func main() {
	flag.Parse()
	g.ParseConfig(*config)

	g.SSHConnAndTransData()

	select {}
}
