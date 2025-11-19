//go:generate goversioninfo ../../versioninfo.json -platform-specific=true
package main

import (
	"flag"
	"fmt"
	"os"

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

	// Handle Lumos
	if *nightFlag != "" {
		hasOperation = true
		if err := handleLumos(*nightFlag); err != nil {
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

func handleLumos(state string) error {
	nl := n.NewLumos()

	switch state {
	case "on":
		if err := nl.Enable(); err != nil {
			return err
		}
		fmt.Println("Lumos enabled")
	case "off":
		if err := nl.Disable(); err != nil {
			return err
		}
		fmt.Println("Lumos disabled")
	case "toggle":
		if err := nl.Toggle(); err != nil {
			return err
		}
		fmt.Println("Lumos toggled")
	default:
		return fmt.Errorf("invalid night light state: %s (must be 'on', 'off', or 'toggle')", state)
	}
	return nil
}

func printHelp() {
	fmt.Println("Usage: lumos [--hdr on|off|toggle] [--gamma <0-100>] [--night on|off|toggle]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --hdr on|off|toggle  Control HDR")
	fmt.Println("  --gamma 0-100        Set gamma level")
	fmt.Println("  --night on|off|toggle Control night light")
	fmt.Println("  --help               Show help")
	fmt.Println("  --version            Show version")
}
