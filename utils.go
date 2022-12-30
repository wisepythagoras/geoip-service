package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func sliceContains[T any](arr []T, thing T) bool {
	for _, v := range arr {
		if reflect.DeepEqual(v, thing) {
			return true
		}
	}

	return false
}

// ParseIPList parses a list of IP addresses and CIDR ranges into two lists which could be used
// as a whitelist or blacklist.
func ParseIPList(file *os.File) ([]*net.IPNet, []net.IP, error) {
	scanner := bufio.NewScanner(file)
	ipRanges := []*net.IPNet{}
	ipAddresses := []net.IP{
		net.ParseIP("127.0.0.1"),
	}

	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile(`(#[\s\S]+)`)
		line = strings.Trim(re.ReplaceAllLiteralString(line, ""), " ")

		if len(line) == 0 {
			continue
		}

		isCidrRange := strings.Contains(line, "/")

		if isCidrRange {
			_, ipRange, err := net.ParseCIDR(line)

			if err != nil {
				return nil, nil, err
			}

			ipRanges = append(ipRanges, ipRange)
		} else {
			ipAddresses = append(ipAddresses, net.ParseIP(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return ipRanges, ipAddresses, nil
}

// ParseDNSServerList parses a file which has a DNS server on each line. The format is:
// 1.1.1.1:53 for the DNS server. Comments are allowed on the same line preceeded by #.
func ParseDNSServerList(file *os.File) ([]string, error) {
	scanner := bufio.NewScanner(file)
	dnsServerList := []string{}

	for scanner.Scan() {
		line := scanner.Text()

		re := regexp.MustCompile(`(#[\s\S]+)`)
		line = strings.Trim(re.ReplaceAllLiteralString(line, ""), " ")

		if len(line) == 0 {
			continue
		}

		if !strings.Contains(line, ":") {
			line = fmt.Sprintf("%s:53", line)
		}

		dnsServerList = append(dnsServerList, line)
	}

	return dnsServerList, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	if err != nil || os.IsNotExist(err) {
		return false
	}

	return true
}
