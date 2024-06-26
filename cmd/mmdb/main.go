package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/btwiuse/mmdb"
	"github.com/btwiuse/tags"
	"github.com/maxmind/mmdbinspect/pkg/mmdbinspect"
)

func usage() {
	fmt.Printf(
		"Usage: %s [-include-aliased-networks] -db path/to/db -db path/to/other/db 130.113.64.30/24 0:0:0:0:0:ffff:8064:a678\n", //nolint: lll
		os.Args[0],
	)
	flag.PrintDefaults()
	fmt.Print(`
Any additional arguments passed are assumed to be networks to look up.  If an
address range is not supplied, /32 will be assumed for ipv4 addresses and /128
will be assumed for ipv6 addresses.
`)
}

func main() {
	var dbs tags.CommaSeparatedStrings
	DefaultDBs, err := mmdb.EnsureLatestDBFiles()
	if err != nil {
		log.Fatal(err)
	}

	flag.Var(&dbs, "db", "Path to an mmdb file. You may pass this arg more than once.")
	includeAliasedNetworks := flag.Bool(
		"include-aliased-networks", false,
		"Include aliased networks (e.g. 6to4, Teredo). This option may cause IPv4 networks to be listed more than once via aliases.", //nolint: lll
	)

	flag.Usage = usage
	flag.Parse()

	// Any remaining arguments (not passed via flags) should be networks
	network := flag.Args()

	if len(network) == 0 {
		fmt.Println("You must provide at least one network address")
		usage()
		os.Exit(1)
	}

	if len(dbs) == 0 {
		dbs = DefaultDBs
	}

	records, err := mmdbinspect.AggregatedRecords(network, dbs, *includeAliasedNetworks)
	if err != nil {
		log.Fatal(err)
	}

	// anonymize the record paths
	for i, record := range records.([]mmdbinspect.RecordSet) {
		record.Database = path.Base(record.Database)
		records.([]mmdbinspect.RecordSet)[i] = record
	}

	json, err := mmdbinspect.RecordToString(records)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", json)
}
