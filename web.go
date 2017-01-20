package main

// Slack outgoing webhooks are handled here. Requests come in and are run through
// the markov chain to generate a response, which is sent back to Slack.
//
// Create an outgoing webhook in your Slack here:
// https://my.slack.com/services/new/outgoing-webhook

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

type WebhookResponse struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

func init() {
	botStatus = "enabled"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		incomingText := r.PostFormValue("text")
		if incomingText != "" && r.PostFormValue("user_id") != "" {
			text := parseText(incomingText)

			// Keep the control word out of our markov chain
			if strings.HasPrefix(text, botControlWord) && r.PostFormValue("user_name") != botAPIName {
				log.Printf("Handling incoming request: %s", text)

				// Strip the keyword from our command
				command := strings.TrimSpace(strings.Replace(text, botControlWord, "", 1))
				w.Write(botControl(command))
			} else {
				if rand.Intn(100) <= responseChance || seeMyName(text) {
					w.Write(generateResponse(botUsername, markovChain.Generate(numWords), true))
				}
			}
		}
	})
}

func StartServer(port int) {
	log.Printf("Starting HTTP server on %d", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
