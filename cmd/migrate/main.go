package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leenwood/market-auth-service/internal/app/migrate"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up | down | status | reset")
	flag.Parse()

	if err := migrate.Run(*direction); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
