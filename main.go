package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var rspamdURL = flag.String("url", "http://127.0.0.1:11333", "rspamd control url")

type rspamdResponse struct {
	Action string
	Score  float32
}

func rspamdRequest(url string, body io.Reader) (rspamdResponse, error) {
	decodedResp := rspamdResponse{}

	// protocol: https://rspamd.com/doc/architecture/protocol.html
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/checkv2", url), body)
	if err != nil {
		return decodedResp, err
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
		return decodedResp, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&decodedResp); err != nil {
		return decodedResp, err
	}

	return decodedResp, nil
}

func main() {
	flag.Parse()
	response, err := rspamdRequest(*rspamdURL, bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("X-Spam-Action: %s\n", response.Action)
	fmt.Printf("X-Spam-Score: %f\n", response.Score)

	switch response.Action {
	case "reject":
		// hard reject
		os.Exit(20)
	case "soft reject":
		// greylist
		os.Exit(75)
	}
}
