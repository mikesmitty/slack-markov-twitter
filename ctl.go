package main

import "fmt"

func botControl(command string) []byte {
	switch command {
	case "status":
		status := fmt.Sprintf("%s is currently %s", botUsername, botStatus)
		return generateResponse(botUsername, status, false)
	case "shutdown", "stop":
		if botStatus == "disabled" {
			status := fmt.Sprintf("%s is already disabled", botUsername)
			return generateResponse(botUsername, status, false)
		} else {
			botStatus = "disabled"
			status := fmt.Sprintf("%s is shutting down", botUsername)
			return generateResponse(botUsername, status, false)
		}
	case "start":
		if botStatus == "enabled" {
			status := fmt.Sprintf("%s is already running", botUsername)
			return generateResponse(botUsername, status, false)
		} else {
			botStatus = "enabled"
			status := fmt.Sprintf("%s is starting up", botUsername)
			return generateResponse(botUsername, status, false)
		}
	default:
		status := fmt.Sprintf("%s doesn't know that command |%s|", botUsername, command)
		return generateResponse(botUsername, status, false)
	}
}
