package main

import (
    "strings"
    "regexp"
    "net"
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

