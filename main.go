package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strings"
)

var rspamdURL = flag.String("url", "http://127.0.0.1:11333", "rspamd control url")
var extendedStats = flag.Bool("extended", false, "report extended stats in X-Spamd-Result header")

type rspamdSymbol struct {
	Score   float32
	Options []string
}

type rspamdResponse struct {
	Action        string
	Score         float32
	RequiredScore float32 `json:"required_score"`
	Symbols       map[string]rspamdSymbol
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

	resp, err := client.Do(req)
	if err != nil {
		return decodedResp, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&decodedResp); err != nil {
		return decodedResp, err
	}

	return decodedResp, nil
}

func formatSymbol(name string, symbol rspamdSymbol) string {
	return fmt.Sprintf("%s(%.2f)[%s]", name, symbol.Score, strings.Join(symbol.Options, ", "))
}

func formatSymbols(symbols map[string]rspamdSymbol) []string {
	var symbolStrings = make([]string, len(symbols))
	i := 0
	for key, value := range symbols {
		symbolStrings[i] = formatSymbol(key, value)
		i++
	}
	return symbolStrings
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
	if *extendedStats {
		fmt.Printf("X-Spamd-Result: default: False [%.2f / %.2f];\n", response.Score, response.RequiredScore)
		symbolStrings := formatSymbols(response.Symbols)
		for i, symbolString := range symbolStrings {
			var lineEnd string
			if i >= len(symbolStrings)-1 {
				lineEnd = "\n"
			} else {
				lineEnd = ";\n"
			}
			fmt.Printf("\t%s%s", symbolString, lineEnd)
		}
	} else {
		fmt.Printf("X-Spam-Score: %.2f\n", response.Score)
	}
}
