package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/api/dns/v1"
)

const (
	ipURL = "https://ipecho.net/plain"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.DurationVar(
		&http.DefaultClient.Timeout,
		"timeout",
		5*time.Second,
		"Timeout for outgoing HTTP Requests",
	)

	flag.Parse()
}

func main() {
	//
	// Determine own IP from external service
	//
	res, err := http.Get(ipURL)
	check(err)

	defer res.Body.Close()
	ip, err := ioutil.ReadAll(res.Body)
	check(err)
	targetIP := string(ip)

	//
	// Determine current DNS records
	//
	domain := mustGetEnv("DOMAIN")

	dnsIPs, err := net.LookupHost(domain)
	check(err)

	if len(dnsIPs) == 1 && dnsIPs[0] == targetIP {
		log.Printf("DNS record matches desired state (%s), exiting", ip)
		return
	}

	log.Printf("Got '%s' instead of '[%s]', checking with DNS provider",
		dnsIPs, ip)

	//
	// Check with DNS provider (if there's TTL lag)
	//
	var (
		ctx     = context.Background()
		project = mustGetEnv("GOOGLE_PROJECT")
		zone    = mustGetEnv("GOOGLE_ZONE")
	)

	dnsService, err := dns.NewService(ctx)
	rrlist, err := dnsService.ResourceRecordSets.
		List(project, zone).
		Name(domain + ".").
		Do()
	check(err)

	if len(rrlist.Rrsets) != 1 {
		log.Fatalf("Could not find correct DNS record, got %#v", rrlist)
	}
	providerZone := rrlist.Rrsets[0]
	providerIPs := providerZone.Rrdatas

	if len(providerIPs) == 1 && providerIPs[0] == targetIP {
		log.Printf("Provider records match desired state (%s), exiting", ip)
		return
	}

	log.Printf("Provider has '%s' instead of '[%s]', updating record",
		dnsIPs, ip)

	change := dnsService.Changes.Create(project, zone, &dns.Change{
		Additions: []*dns.ResourceRecordSet{{
			Kind:    providerZone.Kind,
			Name:    providerZone.Name,
			Rrdatas: []string{targetIP},
			Ttl:     providerZone.Ttl,
			Type:    providerZone.Type,
		}},
		Deletions: []*dns.ResourceRecordSet{providerZone},
	})
	check(change.Do())
	log.Println("Updated provider record to new IP", targetIP)
}

func check(errs ...interface{}) {
	for _, err := range errs {
		if err, ok := err.(error); ok && err != nil {
			log.Panicln(err)
		}
	}
}

func mustGetEnv(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	log.Panicf("%s is not set", name)
	return ""
}
