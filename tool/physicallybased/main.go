// physicallybased generates the `physical` GSL library package from a
// snapshot of the Physically Based materials database
// (https://physicallybased.info, CC0).
//
// usage: physicallybased [-fetch] [-snapshot materials.json] [-o physical.gsl]
//
// -fetch refreshes the snapshot from the API before generating.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const apiURL = "https://api.physicallybased.info/v2/materials"

func main() {
	fetch := flag.Bool("fetch", false, "refresh the snapshot from "+apiURL)
	snap := flag.String("snapshot", "materials.json", "snapshot path")
	out := flag.String("o", "../../stdlib/material/bulk/physical/physical.gsl", "output .gsl path")
	flag.Parse()
	if err := run(*fetch, *snap, *out); err != nil {
		fmt.Fprintln(os.Stderr, "physicallybased:", err)
		os.Exit(1)
	}
}

func run(fetch bool, snapPath, outPath string) error {
	if fetch {
		c := http.Client{Timeout: 30 * time.Second}
		resp, err := c.Get(apiURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GET %s: %s", apiURL, resp.Status)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if _, _, err := parseSnapshot(b); err != nil {
			return err
		}
		if err := os.WriteFile(snapPath, b, 0644); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(snapPath)
	if err != nil {
		return err
	}
	s, dropped, err := parseSnapshot(b)
	if err != nil {
		return err
	}
	gsl, err := generate(s, dropped)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(gsl), 0644)
}
