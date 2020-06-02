package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var rspamdURL = flag.String("url", "http://127.0.0.1:11333", "rspamd control url")

type rspamdResponse struct {
	Action string
	Score  float32
}

func rspamdRequest() {
	// protocol: https://rspamd.com/doc/architecture/protocol.html
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/checkv2", *rspamdURL), bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Pass", "all")
	// add some metadata as headers
	auth, authPresent := os.LookupEnv("AUTH_AS")
	if authPresent {
		req.Header.Add("User", auth)
	}
	req.Header.Add("Ip", os.Getenv("REMOTE_ADDR"))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	decodedResp := rspamdResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&decodedResp); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("X-Spam-Action: %s\n", decodedResp.Action)
	fmt.Printf("X-Spam-Score: %f\n", decodedResp.Score)

	switch decodedResp.Action {
	case "reject":
		// hard reject
		os.Exit(20)
	case "soft reject":
		// greylist
		os.Exit(75)
	}

}

func main() {
	flag.Parse()
	rspamdRequest()
}
