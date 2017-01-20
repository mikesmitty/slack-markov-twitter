package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	messageRegex *regexp.Regexp
)

func init() {
	messageRegex = regexp.MustCompile(`<([^>]+)>`)
}

func parseText(text string) string {
	matches := messageRegex.FindAllStringSubmatch(text, -1)
	for _, matches2 := range matches {

		if strings.HasPrefix(matches2[1], "http") || strings.HasPrefix(matches2[1], "mailto") {
			text = strings.Replace(text, matches2[0], "", -1)

		} else if strings.HasPrefix(matches2[1], "@U") {
			parts := strings.SplitN(matches2[1], "|", 2)

			if len(parts) == 2 {
				text = strings.Replace(text, matches2[0], "@"+parts[1], -1)
			} else {
				text = strings.Replace(text, matches2[0], "", -1)
			}

		} else if strings.HasPrefix(matches2[1], "@") {
			text = strings.Replace(text, matches2[0], matches2[1], -1)

		} else if strings.HasPrefix(matches2[1], "#") {
			parts := strings.SplitN(matches2[1], "|", 2)

			if len(parts) == 2 {
				text = strings.Replace(text, matches2[0], "#"+parts[1], -1)
			} else {
				text = strings.Replace(text, matches2[0], "", -1)
			}

		}
	}

	text = strings.TrimSpace(text)

	text = strings.Replace(text, "&lt;", "<", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&amp;", "&", -1)

	return text
}

func generateResponse(username, text string, twitter bool) []byte {
	var response WebhookResponse
	response.Username = username
	response.Text = text
	log.Printf("Sending response: %s", response.Text)

	b, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	//if twitter && twitterClient != nil {
	//	log.Printf("Tweeting: %s", response.Text)
	//	twitterClient.Post(response.Text)
	//}

	time.Sleep(time.Duration(responseTimeout) * time.Second)
	return b
}

func seeMyName(text string) bool {
	text = strings.ToLower(text)

	if alwaysReply {
		return strings.Contains(text, botUsernameLC)
	} else {
		return strings.HasPrefix(text, botUsernameLC)
	}
}
