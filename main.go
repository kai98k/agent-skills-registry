package main

import "github.com/liuyukai/agentskills/cmd"

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, buildTime)
	cmd.Execute()
}
