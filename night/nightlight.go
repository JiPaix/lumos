package night

import (
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	STATE_KEY_PATH    = `Software\Microsoft\Windows\CurrentVersion\CloudStore\Store\DefaultAccount\Current\default$windows.data.bluelightreduction.bluelightreductionstate\windows.data.bluelightreduction.bluelightreductionstate`
	SETTINGS_KEY_PATH = `Software\Microsoft\Windows\CurrentVersion\CloudStore\Store\DefaultAccount\Current\default$windows.data.bluelightreduction.settings\windows.data.bluelightreduction.settings`
	MIN_KELVIN        = 1200 // Maximum warmth (100% strength)
	MAX_KELVIN        = 6500 // Neutral (0% strength)
)

// Lumos represents a controller for Windows night lights feature
type Lumos struct {
	stateKey    string
	settingsKey string
}

// NewLumos creates a new Lumos instance
func NewLumos() *Lumos {
	return &Lumos{
		stateKey:    STATE_KEY_PATH,
		settingsKey: SETTINGS_KEY_PATH,
	}
}

// Supported checks if Lumos is supported on this system
func (nl *Lumos) Supported() bool {
	_, err := nl.getStateData()
	return err == nil
}

// getStateData retrieves the Data value from the state registry key
func (nl *Lumos) getStateData() ([]byte, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, nl.stateKey, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	data, _, err := key.GetBinaryValue("Data")
	return data, err
}

// getSettingsData retrieves the Data value from the settings registry key
func (nl *Lumos) getSettingsData() ([]byte, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, nl.settingsKey, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	data, _, err := key.GetBinaryValue("Data")
	return data, err
}

// setStateData writes the Data value to the state registry key
func (nl *Lumos) setStateData(data []byte) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, nl.stateKey, registry.WRITE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetBinaryValue("Data", data)
}

// setSettingsData writes the Data value to the settings registry key
func (nl *Lumos) setSettingsData(data []byte) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, nl.settingsKey, registry.WRITE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetBinaryValue("Data", data)
}

// Enabled checks if Lumos is currently enabled
func (nl *Lumos) Enabled() (bool, error) {
	if !nl.Supported() {
		return false, errors.New("night light not supported")
	}

	data, err := nl.getStateData()
	if err != nil {
		return false, err
	}

	if len(data) < 19 {
		return false, errors.New("invalid state data length")
	}

	return data[18] == 0x15, nil // 21 in decimal
}

// Enable turns on Lumos
func (nl *Lumos) Enable() error {
	enabled, err := nl.Enabled()
	if err != nil {
		return err
	}

	if !enabled {
		return nl.Toggle()
	}
	return nil
}

// Disable turns off Lumos
func (nl *Lumos) Disable() error {
	enabled, err := nl.Enabled()
	if err != nil {
		return err
	}

	if enabled {
		return nl.Toggle()
	}
	return nil
}

// Toggle toggles Lumos on/off
func (nl *Lumos) Toggle() error {
	enabled, err := nl.Enabled()
	if err != nil {
		return err
	}

	data, err := nl.getStateData()
	if err != nil {
		return err
	}

	var newData []byte

	if enabled {
		// Disable: create a smaller array and copy data with gaps
		newData = make([]byte, 41)
		copy(newData[0:22], data[0:22])
		copy(newData[23:], data[25:43])
		newData[18] = 0x13
	} else {
		// Enable: create a larger array and copy data with gaps
		newData = make([]byte, 43)
		copy(newData[0:22], data[0:22])
		copy(newData[25:], data[23:41])
		newData[18] = 0x15
		newData[23] = 0x10
		newData[24] = 0x00
	}

	// Increment timestamp bytes
	for i := 10; i < 15; i++ {
		if newData[i] != 0xff {
			newData[i]++
			break
		}
	}

	return nl.setStateData(newData)
}

// GetStrength returns the current Lumos strength as a percentage (0-100)
func (nl *Lumos) GetStrength() (float64, error) {
	if !nl.Supported() {
		return 0, errors.New("night light not supported")
	}

	data, err := nl.getSettingsData()
	if err != nil {
		return 0, err
	}

	if len(data) < 0x25 {
		return 0, errors.New("invalid settings data length")
	}

	kelvin := nl.bytesToKelvin(data[0x23], data[0x24])
	return nl.kelvinToPercentage(kelvin), nil
}

// SetStrength sets the Lumos strength (0-100)
func (nl *Lumos) SetStrength(percentage float64) error {
	if !nl.Supported() {
		return errors.New("night light not supported")
	}

	// Clamp percentage between 0-100
	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

	// Convert percentage to kelvin
	kelvin := nl.percentageToKelvin(percentage)

	data, err := nl.getSettingsData()
	if err != nil {
		return err
	}

	if len(data) < 0x25 {
		return errors.New("invalid settings data length")
	}

	// Calculate bytes using the PowerShell script's formula
	tempHi := byte(kelvin / 64)
	tempLo := byte(((kelvin - float64(tempHi)*64) * 2) + 128)

	// Update strength bytes (indices 0x23, 0x24)
	data[0x23] = tempLo
	data[0x24] = tempHi

	// Update timestamp bytes
	for i := 10; i < 15; i++ {
		if data[i] != 0xff {
			data[i]++
			break
		}
	}

	return nl.setSettingsData(data)
}

// bytesToKelvin converts registry bytes to kelvin temperature
func (nl *Lumos) bytesToKelvin(loTemp, hiTemp byte) float64 {
	// Convert bytes back to kelvin using the inverse of the PowerShell formula
	return float64(hiTemp)*64 + (float64(loTemp)-128)/2
}

// kelvinToPercentage converts kelvin temperature to percentage strength
func (nl *Lumos) kelvinToPercentage(kelvin float64) float64 {
	// Inverse linear mapping from kelvin to percentage
	return 100 - ((kelvin-MIN_KELVIN)/(MAX_KELVIN-MIN_KELVIN))*100
}

// percentageToKelvin converts percentage strength to kelvin temperature
func (nl *Lumos) percentageToKelvin(percentage float64) float64 {
	// Linear mapping from percentage to kelvin
	return MAX_KELVIN - (percentage/100)*(MAX_KELVIN-MIN_KELVIN)
}

// Helper functions for hex/byte conversion (not used in main logic but kept for reference)
func hexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

func bytesToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// Example usage
func main() {
	nl := NewLumos()

	if !nl.Supported() {
		fmt.Println("Lumos is not supported on this system")
		return
	}

	// Check current status
	enabled, err := nl.Enabled()
	if err != nil {
		fmt.Printf("Error checking Lumos status: %v\n", err)
		return
	}

	fmt.Printf("Lumos is currently: %s\n", map[bool]string{true: "ENABLED", false: "DISABLED"}[enabled])

	// Get current strength
	strength, err := nl.GetStrength()
	if err != nil {
		fmt.Printf("Error getting strength: %v\n", err)
		return
	}
	fmt.Printf("Current strength: %.1f%%\n", strength)

	// Example: Set strength to 75%
	err = nl.SetStrength(75)
	if err != nil {
		fmt.Printf("Error setting strength: %v\n", err)
	} else {
		fmt.Println("Strength set to 75%")
	}

	// Example: Enable Lumos
	err = nl.Enable()
	if err != nil {
		fmt.Printf("Error enabling Lumos: %v\n", err)
	} else {
		fmt.Println("Lumos enabled")
	}
}
