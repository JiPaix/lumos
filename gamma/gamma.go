package gamma

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procGetDC              = user32.NewProc("GetDC")
	procReleaseDC          = user32.NewProc("ReleaseDC")
	procCreateDC           = gdi32.NewProc("CreateDCW")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
	procSetGammaRamp       = gdi32.NewProc("SetDeviceGammaRamp")
	procEnumDisplayDevices = user32.NewProc("EnumDisplayDevicesW")
)

// DISPLAY_DEVICE structure
type displayDevice struct {
	cb           uint32
	DeviceName   [32]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

// GammaRamp represents the gamma ramp structure
type GammaRamp struct {
	Red   [256]uint16
	Green [256]uint16
	Blue  [256]uint16
}

// SetGamma sets the screen gamma with brightness (0-100) for ALL displays
func SetGamma(brightness int) error {
	if brightness < 0 || brightness > 100 {
		return errors.New("brightness must be between 0 and 100")
	}

	// Calculate factors based on PowerShell logic
	brightnessFactor := 0.5 + (float64(brightness)/100.0)*0.5
	contrast := 120.0 - (0.2 * float64(brightness))
	contrastFactor := contrast / 100.0

	// Create gamma ramp
	var ramp GammaRamp

	for i := range 256 {
		x := float64(i) / 255.0

		// Apply contrast adjustment
		x = ((x - 0.5) * contrastFactor) + 0.5

		// Clamp to [0, 1]
		if x < 0 {
			x = 0
		}
		if x > 1 {
			x = 1
		}

		// Apply brightness and convert to 16-bit value
		val := uint16(x * 65535 * brightnessFactor)

		ramp.Red[i] = val
		ramp.Green[i] = val
		ramp.Blue[i] = val
	}

	// Apply the gamma ramp to ALL displays
	return setDeviceGammaRampForAllDisplays(&ramp)
}

// setDeviceGammaRampForAllDisplays sets gamma ramp for all active displays
func setDeviceGammaRampForAllDisplays(ramp *GammaRamp) error {
	displays := getActiveDisplayDevices()
	var lastError error
	successCount := 0

	// Convert "DISPLAY" once outside the loop since it's constant
	displayPtr, err := syscall.UTF16PtrFromString("DISPLAY")
	if err != nil {
		return errors.New("failed to convert DISPLAY constant")
	}

	for _, deviceName := range displays {
		// Use UTF16PtrFromString for deviceName with proper error handling
		deviceNamePtr, err := syscall.UTF16PtrFromString(deviceName)
		if err != nil {
			// Skip this monitor if conversion fails
			continue
		}

		hdc, _, _ := procCreateDC.Call(
			uintptr(unsafe.Pointer(displayPtr)),
			uintptr(unsafe.Pointer(deviceNamePtr)),
			0, 0,
		)

		if hdc != 0 {
			err := setGammaWithHDC(hdc, ramp)
			procDeleteDC.Call(hdc) // Clean up DC

			if err != nil {
				lastError = err
			} else {
				successCount++
			}
		}
	}

	// If no specific displays worked, fallback to primary display
	if successCount == 0 {
		hdc, _, _ := procGetDC.Call(0)
		if hdc == 0 {
			if lastError != nil {
				return lastError
			}
			return errors.New("failed to set gamma on any display")
		}
		defer procReleaseDC.Call(0, hdc)

		return setGammaWithHDC(hdc, ramp)
	}

	if lastError != nil && successCount == 0 {
		return lastError
	}

	return nil
}

// setGammaWithHDC sets gamma ramp using a specific HDC
func setGammaWithHDC(hdc uintptr, ramp *GammaRamp) error {
	ret, _, err := procSetGammaRamp.Call(hdc, uintptr(unsafe.Pointer(&ramp.Red[0])))
	if ret == 0 {
		return errors.New("failed to set gamma ramp: " + err.Error())
	}
	return nil
}

// getActiveDisplayDevices returns a list of active display device names
func getActiveDisplayDevices() []string {
	var devices []string
	var dd displayDevice
	dd.cb = uint32(unsafe.Sizeof(dd))

	index := 0
	for {
		// EnumDisplayDevices with nullptr for device name to enumerate all devices
		ret, _, _ := procEnumDisplayDevices.Call(
			0,
			uintptr(index),
			uintptr(unsafe.Pointer(&dd)),
			0,
		)

		if ret == 0 {
			break
		}

		// Check if this is an active display device (attached to desktop)
		active := (dd.StateFlags & 0x00000001) != 0 // DISPLAY_DEVICE_ACTIVE

		// Only include active displays that are attached to desktop
		if active {
			deviceName := syscall.UTF16ToString(dd.DeviceName[:])
			devices = append(devices, deviceName)
		}

		index++
	}

	return devices
}

// ResetGamma resets the gamma to default (brightness 100) for all displays
func ResetGamma() error {
	return SetGamma(100)
}

// GetDisplayCount returns the number of available active displays
func GetDisplayCount() int {
	return len(getActiveDisplayDevices())
}
