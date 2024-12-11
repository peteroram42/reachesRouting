/**
	Net utils used by reachesRouting
*/

package reachesRouting

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"bytes"
)

//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
    start net.IP
    end net.IP
}

// private IPv4 ranges
var privateRanges = []ipRange{
    ipRange{
        start: net.ParseIP("10.0.0.0"),
        end:   net.ParseIP("10.255.255.255"),
    },
    ipRange{
        start: net.ParseIP("100.64.0.0"),
        end:   net.ParseIP("100.127.255.255"),
    },
    ipRange{
        start: net.ParseIP("172.16.0.0"),
        end:   net.ParseIP("172.31.255.255"),
    },
    ipRange{
        start: net.ParseIP("192.0.0.0"),
        end:   net.ParseIP("192.0.0.255"),
    },
    ipRange{
        start: net.ParseIP("192.168.0.0"),
        end:   net.ParseIP("192.168.255.255"),
    },
    ipRange{
        start: net.ParseIP("198.18.0.0"),
        end:   net.ParseIP("198.19.255.255"),
    },
}

// Get the IP address out of a request, taking into account the
// "True-Client-IP". Works for ipv6 as well.
func getIpFromRequest(r *http.Request) (ip net.IP, err error) {

	var ip_string string

	// allow test IPs
	if test_ip := r.URL.Query().Get("testip"); test_ip != "" {
		ip_string = test_ip
	}

	// take Akamai's True-Client-IP (if possible)
	if ip_string == "" {
		ip_string = r.Header.Get("True-Client-IP")	
	} 

	// Now going to check several headers that could contain the IP, marching
	// from left to right to get earliest IP used and ignoring private or 
	// internal IPs, so we get their first public internet contact point
	// structs, 
	if ip_string == "" {
		ip_string = getIPAdress(r)
	}

	if ip_string == "" {
		// uable to get the IP
		err = errors.New("Unable to determine IP address string")
		return
	}

	// get the visitor's IP
	ip = net.ParseIP(ip_string)
	if ip == nil {
		err = errors.New("Unable to determine IP address")
		return
	}

	return
}


// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
    // strcmp type byte comparison
    if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
        return true
    }
    return false
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
    // my use case is only concerned with ipv4 atm
    if ipCheck := ipAddress.To4(); ipCheck != nil {
        // iterate over all our ranges
        for _, r := range privateRanges {
            // check if this ip is in a private range
            if inRange(r, ipAddress){
                return true
            }
        }
    }
    return false
}

func getIPAdress(r *http.Request) string {
    var ipAddress string
    for _, h := range []string{"X-Forwarded-For", "X-Real-Ip", "X-ProxyUser-Ip"} {
        for _, ip := range strings.Split(r.Header.Get(h), ",") {
            // header can contain spaces too, strip those out.
            ip = strings.TrimSpace(ip)
            realIP := net.ParseIP(ip)
            if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
                // bad address, go to next
                continue
            } else {
                ipAddress = ip
                goto Done
            }
        }
    }
    Done:
    return ipAddress
}
