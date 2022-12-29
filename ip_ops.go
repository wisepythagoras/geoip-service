package main

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
)

func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

func IsValidIP(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

	if re.MatchString(ipAddress) {
		return true
	}

	return IsIPv6(ipAddress)
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
