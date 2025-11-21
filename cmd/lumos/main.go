//go:generate goversioninfo ../../versioninfo.json -platform-specific=true
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/jipaix/lumos/gamma"
	"github.com/jipaix/lumos/hdr"
	n "github.com/jipaix/lumos/night"
)

const version = "1.0"

func main() {
	// Define flags
	hdrFlag := flag.String("hdr", "", "Set HDR state (on/off/toggle)")
	gammaFlag := flag.Int("gamma", -1, "Set gamma percentage (0-100)")
	nightFlag := flag.String("night", "", "Set night light state (on/off/toggle)")
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")

	// Custom usage function
	flag.Usage = printHelp

	// Parse flags
	flag.Parse()

	// Handle help and version flags
	if *helpFlag || len(os.Args) == 1 {
		printHelp()
		return
	}

	if *versionFlag {
		fmt.Println(version)
		return
	}

	// Execute commands based on flags
	var hasOperation bool

	// Handle HDR
	if *hdrFlag != "" {
		hasOperation = true
		if err := handleHDR(*hdrFlag); err != nil {
			fmt.Printf("Error setting HDR: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle Gamma
	if *gammaFlag != -1 {
		hasOperation = true
		if err := handleGamma(*gammaFlag); err != nil {
			fmt.Printf("Error setting gamma: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle Night Light
	if *nightFlag != "" {
		hasOperation = true
		if err := handleNightLight(*nightFlag); err != nil {
			fmt.Printf("Error setting night light: %v\n", err)
			os.Exit(1)
		}
	}

	// If no valid operations were performed, show help
	if !hasOperation {
		printHelp()
	}
}

func handleHDR(hdrState string) error {
	hdrCtrl := hdr.NewHDR()

	// Check if HDR is supported
	if !hdrCtrl.IsHDRSupported() {
		return fmt.Errorf("HDR is not supported on this system or requires administrator privileges")
	}

	switch hdrState {
	case "on":
		if err := hdrCtrl.Enable(); err != nil {
			return fmt.Errorf("HDR enable failed: %v", err)
		}
		fmt.Println("HDR enabled on all compatible displays")
	case "off":
		if err := hdrCtrl.Disable(); err != nil {
			return fmt.Errorf("HDR disable failed: %v", err)
		}
		fmt.Println("HDR disabled on all compatible displays")
	case "toggle":
		if err := hdrCtrl.Toggle(); err != nil {
			return fmt.Errorf("HDR toggle failed: %v", err)
		}
		fmt.Println("HDR toggled on all compatible displays")
	default:
		return fmt.Errorf("invalid HDR state: %s (must be 'on', 'off', or 'toggle')", hdrState)
	}
	return nil
}

func handleGamma(percentage int) error {
	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("gamma percentage must be between 0 and 100, got %d", percentage)
	}

	if err := gamma.SetGamma(percentage); err != nil {
		return err
	}
	fmt.Printf("Gamma set to %d%% on all displays\n", percentage)
	return nil
}

func handleNightLight(state string) error {
	nl := n.NewLumos()

	switch state {
	case "on":
		if err := nl.Enable(); err != nil {
			return err
		}
		fmt.Println("Night light enabled")
	case "off":
		if err := nl.Disable(); err != nil {
			return err
		}
		fmt.Println("Night light disabled")
	case "toggle":
		if err := nl.Toggle(); err != nil {
			return err
		}
		fmt.Println("Night light toggled")
	default:
		val, err := strconv.ParseFloat(state, 64)
		if err != nil {
			return fmt.Errorf("invalid night light state: %s (must be 'on', 'off', 'toggle', or a percentage like '50')", state)
		}

		if err := nl.SetStrength(val); err != nil {
			return err
		}

		// Check if currently enabled to decide if we need a "hard restart"
		enabled, err := nl.Enabled()
		if err != nil {
			return err
		}

		// If it's currently on, disable it first to force the refresh
		if enabled {
			if err := nl.Disable(); err != nil {
				return err
			}
			time.Sleep(200 * time.Millisecond)
		}

		// Turn it on (or back on)
		if err := nl.Enable(); err != nil {
			return err
		}

		fmt.Printf("Night light strength set to %v%%\n", val)
		return nil
	}
	return nil
}

func printHelp() {
	fmt.Println("Usage: lumos [--hdr on|off|toggle] [--gamma <0-100>] [--night on|off|toggle|<0-100>]")
	fmt.Println()
	fmt.Println("Options:")

	// minwidth=0, tabwidth=0, padding=2, padchar=' ', flags=0
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Use \t to separate the flag from the description
	fmt.Fprintln(w, "  --hdr on|off|toggle\tControl HDR")
	fmt.Fprintln(w, "  --gamma <0-100>\tSet gamma level")
	fmt.Fprintln(w, "  --night on|off|toggle|<0-100>\tControl night light")
	fmt.Fprintln(w, "  --help\tShow help")
	fmt.Fprintln(w, "  --version\tShow version")

	w.Flush()
}
