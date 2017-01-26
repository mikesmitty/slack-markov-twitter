// copyright 2012 arne roomann-kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// Reads as much of a user's last 3200 public Tweets as the Twitter API
// returns, and prints each Tweet to a file.
//
// This example respects rate limiting and will wait until the rate limit
// reset time to finish pulling a timeline.
//
// An out of sync clock can make it appear that the reset has passed and
// cause extra requests.  Use the following to synchronize your time:
//     ntpd -q
// Or (use any NTP server):
//     ntpdate ntp.ubuntu.com
//
// If rate limiting happens, you'll see the executable pause until it
// estimates that the limit has reset.  A more robust implementation would
// use a different approach than just sleeping, but this is a simple example.

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
)

func crawlTwitter(aWhile time.Duration, consumerKey, consumerSecret, accessTokenKey, accessTokenSecret, sourceUser string, sinceID uint64) {
	for {
		scrapeTwitter(twitterConsumerKey, twitterConsumerSecret, twitterAccessToken, twitterAccessTokenSecret, twitterSourceUser, markovChain.SinceID)
		time.Sleep(aWhile)
	}
}

func scrapeTwitter(consumerKey, consumerSecret, accessTokenKey, accessTokenSecret, sourceUser string, sinceID uint64) {
	const (
		count   int = 100
		urltmpl     = "/1.1/statuses/user_timeline.json?%v"
		minwait     = time.Duration(10) * time.Second
	)

	var (
		err      error
		client   *twittergo.Client
		req      *http.Request
		resp     *twittergo.APIResponse
		latestID uint64
		max_id   uint64
		query    url.Values
		results  *twittergo.Timeline
	)

	config := &oauth1a.ClientConfig{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
	}
	user := oauth1a.NewAuthorizedConfig(accessTokenKey, accessTokenSecret)
	client = twittergo.NewClient(config, user)

	query = url.Values{}
	query.Set("count", fmt.Sprintf("%v", count))
	query.Set("screen_name", sourceUser)

	total := 0
	for {
		if max_id != 0 {
			query.Set("max_id", fmt.Sprintf("%v", max_id))
		}
		if sinceID != 0 {
			query.Set("since_id", fmt.Sprintf("%v", sinceID))
		}

		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			log.Printf("Could not parse request: %v", err)
			return
		}
		if resp, err = client.SendRequest(req); err != nil {
			log.Printf("Could not send request: %v", err)
			return
		}

		results = &twittergo.Timeline{}
		if err = resp.Parse(results); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v"
				log.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				log.Printf("Problem parsing response: %v", err)
			}
		}

		batch := len(*results)
		if batch == 0 {
			log.Printf("No more results, end of timeline.")
			break
		}

		for _, tweet := range *results {
			text := tweet.Text()
			if text != "" {
				markovChain.Write(text)
			} else {
				log.Printf("Could not save Tweet (ID %d): |%v|", tweet.Id(), text)
			}

			tweetID := tweet.Id()
			if latestID == 0 && tweetID > 0 {
				latestID = tweetID
			}

			max_id = tweetID - 1
			total += 1
		}

		status := fmt.Sprintf("Got %v Tweets", batch)
		if resp.HasRateLimit() {
			status = fmt.Sprintf("%s, %v calls available", status, resp.RateLimitRemaining())
		}
		status = fmt.Sprintf("%s.", status)
		log.Printf(status)
	}
	log.Printf("--------------------------------------------------------")
	log.Printf("Writing %v Tweets to chain", total)

	if latestID != 0 {
		markovChain.SinceID = latestID
		go func() {
			markovChain.Save(stateFile)
		}()
	}
}
