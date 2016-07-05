package main

import "os"

func main() {
	findPluginAndProjectRoot()
	action := os.Getenv(PLUGIN_ACTION_ENV)
	if action == SETUP_ACTION {
		addDefaultPropertiesToProject()
	} else if action == EXECUTION_ACTION {
		createExecutionReport()
	}
}
