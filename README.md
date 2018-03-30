# hover-ddns-go
Hover DDNS script written in Go.

## Installation
```bash
$ go get -u github.com/lakhanmankani/hover-ddns-go
```
or
```bash
$ git clone https://github.com/lakhanmankani/hover-ddns-go.git
```

## Setup
1. Run main.go to generate the config file.
```bash
$ go run main.go
```

2. Edit config.json file. Enter your username and password in the username and password value fields. For every domain create a new object in the "domains" array. Enter the domainID and dnsID in the corresponding value field.

Note: You can find the domainID and dnsID of the domain at https://www.hover.com/api/domains/YOURDOMAIN.COM/dns
## Usage
```bash
$ go run main.go
```
Optional flags:
* ```-v``` Turn on verbose mode.
