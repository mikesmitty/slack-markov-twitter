package main

// Main entry point for the app. Handles command-line options, starts the web
// listener and any import, etc

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	alwaysReply     bool
	botControlWord  string
	botAPIName      string
	botStatus       string
	botUsername     string
	botUsernameLC   string
	chatty          bool
	httpPort        int
	numWords        int
	prefixLen       int
	stateFile       string
	responseChance  int
	responseTimeout int

	twitterTimeout           int
	twitterSourceUser        string
	twitterConsumerKey       string
	twitterConsumerSecret    string
	twitterAccessToken       string
	twitterAccessTokenSecret string
	//twitterClient            *Twitter

	markovChain *Chain
)

func init() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator.
}

func main() {
	// Parse command-line options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ./slack-markov -port=8000\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&alwaysReply, "alwaysReply", false, "Reply whenever the bot sees its name anywhere")
	flag.BoolVar(&chatty, "chatty", false, "Allow to bot to reply to itself")
	flag.IntVar(&httpPort, "port", 8000, "The HTTP port on which to listen")
	flag.IntVar(&numWords, "words", 100, "Maximum number of words in the output")
	flag.IntVar(&prefixLen, "prefix", 2, "Prefix length in words")
	flag.IntVar(&responseChance, "responseChance", 10, "Percent chance to generate a response on each request")
	flag.IntVar(&responseTimeout, "responseTimeout", 2, "Response delay in seconds (to prevent flooding)")
	flag.StringVar(&botControlWord, "botControlWord", "markovctl", "Keyword used to enable/disable the bot")
	flag.StringVar(&botAPIName, "botAPIName", "slackbot", "The name of the bot as received in the API")
	flag.StringVar(&botUsername, "botUsername", "markov-bot", "The name of the bot when it speaks")
	flag.StringVar(&stateFile, "stateFile", "state", "File to use for maintaining our markov chain state")

	flag.IntVar(&twitterTimeout, "twitterTimeout", 60, "Timeout between Twitter scrapes")
	flag.StringVar(&twitterConsumerKey, "twitterConsumerKey", "", "Twitter API key")
	flag.StringVar(&twitterConsumerSecret, "twitterConsumerSecret", "", "Twitter API key secret")
	flag.StringVar(&twitterSourceUser, "twitterUser", "", "Twitter user to markov-ize tweets from")
	flag.StringVar(&twitterAccessToken, "twitterAccessToken", "", "Twitter access token")
	flag.StringVar(&twitterAccessTokenSecret, "twitterAccessTokenSecret", "", "Twitter access token secret")

	//var importDir = flag.String("importDir", "", "The directory of a Slack export")
	//var importChan = flag.String("importChan", "", "Optional channel to limit the import to")

	flag.Parse()

	if httpPort == 0 {
		flag.Usage()
		os.Exit(2)
	}

	markovChain = NewChain(prefixLen) // Initialize a new Chain.

	// Rebuild the markov chain from state
	err := markovChain.Load(stateFile)
	if err != nil {
		//log.Fatal(err)
		log.Printf("Could not load from '%s'. This may be expected.", stateFile)
	} else {
		log.Printf("Loaded previous state from '%s' (%d suffixes).", stateFile, len(markovChain.Chain))
	}

	// Optionally start trolling twitter
	if twitterConsumerKey != "" && twitterConsumerSecret != "" && twitterAccessToken != "" && twitterAccessTokenSecret != "" && twitterSourceUser != "" {
		if twitterTimeout < 1 {
			log.Printf("Can't have a timeout less than one minute. Setting to 60 minutes")
			twitterTimeout = 60
		}
		aWhile := time.Duration(twitterTimeout) * time.Minute
		go crawlTwitter(aWhile, twitterConsumerKey, twitterConsumerSecret, twitterAccessToken, twitterAccessTokenSecret, twitterSourceUser, markovChain.SinceID)
	}

	// Create a lower-case version of the bot name for matching later
	botUsernameLC = strings.ToLower(botUsername)

	// Start the webserver
	StartServer(httpPort)
}
