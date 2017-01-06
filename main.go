package main

import "os"

func main() {
	findProjectRoot()
	action := os.Getenv(pluginActionEnv)
	if action == setupAction {
		addDefaultPropertiesToProject()
	} else if action == executionAction {
		createExecutionReport()
	}
}
