package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

type updateDnsResponse struct {
	Succeeded bool
	Error     string
}

type Config struct {
	Username string
	Password string
	Domains  []Domains
}

type Domains struct {
	DomainId string
	DnsId    string
}

func getCurrentIP() (ipAddress string, err error) {
	url := "http://bot.whatismyipaddress.com"
	resp, err := http.Get(url)
	if err != nil {
		return ipAddress, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)

	}
	return string(body), err

}

func authenticateHover(username string, password string) (auth string) {
	// Login
	authUrl := "https://www.hover.com/signin/auth.json"

	requestContent := bytes.NewBuffer([]byte(`{"username": "` + username + `","password": "` + password + `"}`))

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Post(authUrl, "application/json;charset=UTF-8", requestContent)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	u, err := url.Parse(authUrl)
	if err != nil {
		panic(err)
	}

	for _, cookie := range cookieJar.Cookies(u) {
		if cookie.Name == "hoverauth" {
			auth = cookie.Value
			break
		}
	}
	return
}

func updateDNS(hoverauth string, ipAddress string, domainId string, dnsId string) (err error) {
	dnsUrl := "https://www.hover.com/api/control_panel/dns"

	// Set client
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}

	// Convert string url to http/url type url
	u, err := url.Parse(dnsUrl)
	if err != nil {
		panic(err)
	}

	// Hoverauth cookie
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:  "hoverauth",
		Value: hoverauth,
	}
	cookies = append(cookies, cookie)
	cookieJar.SetCookies(u, cookies)

	// Request
	body :=
		strings.NewReader(fmt.Sprintf(`{"domain":{"id":"%s","dns_records":[{"id":"%s"}]},"fields":{"content":"%s"}}`,
			domainId, dnsId, ipAddress))
	request, err := http.NewRequest("PUT", dnsUrl, body)
	request.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read contents
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Check if succeeded
	jsonResp := updateDnsResponse{}
	json.Unmarshal(contents, &jsonResp)
	if !jsonResp.Succeeded {
		panic(jsonResp.Error)
	}
	return err
}

func main() {
	// Verbose mode flag
	verbosePtr := flag.Bool("v", false, "Turn on verbose mode")
	flag.Parse()

	// Read config.json file
	jsonFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exists create dummy file
			fmt.Println("Config file doesn't exist. Created sample config file.")
			var file, err = os.Create("config.json")
			if err != nil {
				panic(err)
			}

			file.Write([]byte(`{"username":"Username","password":"Password","domains":[{"domainId":"domain-example.com","dnsId":"dns12345678"}]}`))
			defer file.Close()
			os.Exit(0)
		}
		fmt.Println(err)
		panic(err)
	}

	var config Config
	json.Unmarshal(jsonFile, &config)

	// Authenticate
	hoverauth := authenticateHover(config.Username, config.Password)
	if *verbosePtr {
		fmt.Printf("Hoverauth cookie value is %s\n", hoverauth)
	}

	for {
		// Get IP Address
		currentIpAdress, err := getCurrentIP()
		if err != nil {
			fmt.Println(err)
			time.Sleep(60 * time.Second)
			continue
		}
		if *verbosePtr {
			fmt.Printf("Current IP Address is %s\n", currentIpAdress)
		}

		// Update DNS
		for domain := range config.Domains {
			updateDNS(hoverauth, currentIpAdress, config.Domains[domain].DomainId,
				config.Domains[domain].DnsId)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if *verbosePtr {
				fmt.Printf("Updated DNS with domainId %s and dnsID %s to %s\n",
					config.Domains[domain].DomainId,
					config.Domains[domain].DnsId, currentIpAdress)
			}
		}

		time.Sleep(60 * time.Second)
	}
}
