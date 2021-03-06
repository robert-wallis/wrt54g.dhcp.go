// Copyright (C) 2013 Robert Wallis, All Rights Reserved
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
)

var user string
var password string
var hostname string

func main() {
	initFlags()
	flag.Parse()
	url := fmt.Sprint("https://", hostname, "/DHCPTable.asp")
	fmt.Println("Querying", hostname, "for the DHCP client list.")
	html, err := requestHtml(url, user, password)
	if err != nil {
		fmt.Println(err)
		return
	}
	clients, err := FromWRTHtml(string(html))
	if err != nil {
		fmt.Println(err)
		return
	}
	clients.PrintColored()
}

func initFlags() {
	flag.StringVar(&user, "u", "", "the router's admin user")
	flag.StringVar(&password, "p", "admin", "the router's admin password")
	var default_hostname = defaultHostname()
	flag.StringVar(&hostname, "h", default_hostname, "the router's IP address")
	flag.StringVar(&hostname, "ip", default_hostname, "the router's IP address")
}

// find the router ip based on the computer's ip
func defaultHostname() string {
	var interfaces, err = net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return "192.168.1.1"
	}
	// choose ipv4 addresses, because my wrt admin isn't accessable in ipv6
	// the second grouping gets the first three octets
	var router_rex = regexp.MustCompile(`^(((192|10)\.\d+\.\d+\.)\d+)[^\d\.]?`)
	for _, inter := range interfaces {
		// make sure the interface is live (up) and not a 'fake' loopback address
		// brain broken, couldn't figure out a single bitwise statement
		if 0 == inter.Flags&net.FlagLoopback && 0 != inter.Flags&net.FlagUp {
			var addrs, err = inter.Addrs()
			if err != nil {
				fmt.Println(err)
				continue
			}
			// any interface could have multiple ip's assigned
			// usually ipv4 and ipv6
			for _, addr := range addrs {
				var ip = addr.String()
				ipmatch := router_rex.FindStringSubmatch(ip)
				if len(ipmatch) > 0 {
					// looks like an ipv4, take the first 3 octets
					// rex match 0 is full string so second grouping is 2
					return ipmatch[2] + "1"
				}
			}
		}
	}
	// default the router's default ip
	// although it should probably error that your're not connected via NAT
	return "192.168.1.1"
}

// get the html from the router's DHCP client list
func requestHtml(url string, user string, password string) (html []byte, err error) {
	// turn of SSL check because the router doesn't have a valid cert
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// make a new client so we can set the basic auth and ignore SSL
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(user, password)
	var res *http.Response
	tries := 0
	for {
		res, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		// keep trying until you get in
		if res.Status[0:3] == "401" {
			tries++
			if tries > 10 {
				return nil, fmt.Errorf("Tried credentials 10 times and always failed.")
			}
			continue
		}
		if res.Status[0:3] != "200" {
			return nil, fmt.Errorf("HTTP status was: %s", res.Status)
		}
		break
	}
	return ioutil.ReadAll(res.Body)
}

// generate Client objects from the JavaScript in a linksys WRT router's DHCP list
func FromWRTHtml(html string) (Clients, error) {
	pattern := `.*table = new Array\(([^\)]+)\).*`
	arraymatch := regexp.MustCompile(pattern).FindStringSubmatch(html)
	if arraymatch == nil || len(arraymatch) < 2 {
		return nil, fmt.Errorf("Unable to find the JS array DHCP client list")
	}
	pattern = `'([^']*)'`
	fields := regexp.MustCompile(pattern).FindAllStringSubmatch(
		arraymatch[1], -1)
	// there are five fields for each client 
	clients := make([]Client, len(fields)/5)
	for f := 0; f < len(fields); f += 5 {
		c := f / 5
		// subfield 1 is the matched part without quotes
		client := &Client{
			fields[f][1],
			fields[f+1][1],
			fields[f+2][1],
			fields[f+3][1],
			fields[f+4][1],
		}
		clients[c] = *client
	}
	return clients, nil
}

type Client struct {
	Name     string
	Ip       string
	Mac      string
	Lease    string
	ClientId string
}
type Clients []Client

// display the clients on a standard CLI
func (c *Clients) PrintPlain() {
	c.Printf("%*s %*s %*s %*s %*s\n")
}

// display the clients on a xtermcolor type CLI
func (c *Clients) PrintColored() {
	c.Printf("\033[32m%*s \033[33m%*s \033[0m%*s %*s %*s\n")
}

// display the clients as you want (usually just use PrintColored)
func (c *Clients) Printf(format string) {
	widths := make([]int, 5)
	for _, client := range *c {
		if widths[0] < len(client.Name) {
			widths[0] = len(client.Name)
		}
		if widths[1] < len(client.Ip) {
			widths[1] = len(client.Ip)
		}
		if widths[2] < len(client.Mac) {
			widths[2] = len(client.Mac)
		}
		if widths[3] < len(client.Lease) {
			widths[3] = len(client.Lease)
		}
		if widths[4] < len(client.ClientId) {
			widths[4] = len(client.ClientId)
		}
	}
	for _, client := range *c {
		fmt.Printf(format,
			widths[0], client.Name,
			widths[1], client.Ip,
			widths[2], client.Mac,
			widths[3], client.Lease,
			widths[4], client.ClientId,
		)
	}
}
