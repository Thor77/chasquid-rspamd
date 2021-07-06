package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"os"
	"strings"
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
	user := os.Getenv("AUTH_AS")
	if user != "" {
		req.Header.Add("User", user)
	}
	remoteAddr := os.Getenv("REMOTE_ADDR")
	var ip string
	if remoteAddr[0] == '[' {
		ip = strings.Split(strings.Split(remoteAddr, "]")[0], "[")[1]
	} else {
		ip = strings.Split(remoteAddr, ":")[0]
	}
	req.Header.Add("Ip", ip)
	addr, err := mail.ParseAddress(os.Getenv("MAIL_FROM"))
	if err != nil {
		return decodedResp, err
	}
	req.Header.Add("From", addr.Address)
	req.Header.Add("Helo", os.Getenv("EHLO_DOMAIN"))

	names, err := net.LookupAddr(ip)
	if err == nil && len(names) > 0 {
		req.Header.Add("Hostname", names[0])
	}

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
		fmt.Println("spam checking failed; try again later")
		os.Exit(75)
	}

	switch response.Action {
	case "reject":
		// hard reject
		fmt.Println("no.")
		os.Exit(20)
	case "soft reject":
		// greylist
		fmt.Println("greylisted, please try again later")
		os.Exit(75)
	case "add header":
		fmt.Println("X-Spam: Yes")
	}

	fmt.Printf("X-Spam-Action: %s\n", response.Action)
	fmt.Printf("X-Spam-Score: %.2f\n", response.Score)
}
