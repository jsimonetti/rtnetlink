package rtnetlink_test

import (
	"log"

	"github.com/jsimonetti/rtnetlink"
)

// List all rules
func Example_listRule() {
	// Dial a connection to the rtnetlink socket
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Request a list of rules
	rules, err := conn.Rule.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, rule := range rules {
		log.Printf("%+v", rule)
	}
}
