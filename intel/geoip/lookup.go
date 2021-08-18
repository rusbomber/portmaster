package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

func getReader(ip net.IP) *maxminddb.Reader {
	isV6 := ip.To4() == nil
	return worker.GetReader(isV6, true)
}

// GetLocation returns Location data of an IP address
func GetLocation(ip net.IP) (*Location, error) {
	db := getReader(ip)
	if db == nil {
		return nil, fmt.Errorf("geoip database not available")
	}
	var record Location
	if err := db.Lookup(ip, &record); err != nil {
		return nil, err
	}
	return &record, nil
}
