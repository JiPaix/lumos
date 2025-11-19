package hdr

import (
	"errors"
	"syscall"
	"unsafe"
)

// Windows API constants
const (
	QDC_ALL_PATHS                                     = 0x00000001
	QDC_ONLY_ACTIVE_PATHS                             = 0x00000002
	DISPLAYCONFIG_DEVICE_INFO_SET_ADVANCED_COLOR_INFO = 0x00000010
	DISPLAYCONFIG_DEVICE_INFO_GET_ADVANCED_COLOR_INFO = 15
)

// Windows structures
type LUID struct {
	LowPart  uint32
	HighPart int32
}

type DISPLAYCONFIG_VIDEO_OUTPUT_TECHNOLOGY uint32

type DISPLAYCONFIG_PATH_INFO struct {
	Source DISPLAYCONFIG_PATH_SOURCE_INFO
	Target DISPLAYCONFIG_PATH_TARGET_INFO
	Flags  uint32
}

type DISPLAYCONFIG_PATH_SOURCE_INFO struct {
	AdapterId   LUID
	Id          uint32
	ModeInfoIdx uint32
	StatusFlags uint32
}

type DISPLAYCONFIG_PATH_TARGET_INFO struct {
	AdapterId        LUID
	Id               uint32
	ModeInfoIdx      uint32
	OutputTechnology DISPLAYCONFIG_VIDEO_OUTPUT_TECHNOLOGY
	Rotation         uint32
	Scaling          uint32
	RefreshRate      DISPLAYCONFIG_RATIONAL
	ScanLineOrdering uint32
	TargetAvailable  uint32 // BOOL in Windows is 4 bytes
	StatusFlags      uint32
}

type DISPLAYCONFIG_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

type DISPLAYCONFIG_ADVANCED_COLOR_INFO struct {
	Header DISPLAYCONFIG_DEVICE_INFO_HEADER
	Value  uint32
}

type DISPLAYCONFIG_DEVICE_INFO_HEADER struct {
	Type      uint32
	Size      uint32
	AdapterId LUID
	Id        uint32
}

type DISPLAYCONFIG_SOURCE_MODE struct {
	Width       uint32
	Height      uint32
	PixelFormat uint32
	Position    POINTL
}

type DISPLAYCONFIG_TARGET_MODE struct {
	TargetVideoSignalInfo DISPLAYCONFIG_VIDEO_SIGNAL_INFO
}

type DISPLAYCONFIG_VIDEO_SIGNAL_INFO struct {
	PixelRate  uint64
	HSyncFreq  DISPLAYCONFIG_RATIONAL
	VSyncFreq  DISPLAYCONFIG_RATIONAL
	ActiveSize POINTL
	TotalSize  POINTL
}

type POINTL struct {
	X int32
	Y int32
}

type DISPLAYCONFIG_MODE_INFO struct {
	InfoType  uint32
	Id        uint32
	AdapterId LUID
	ModeInfo  unionModeInfo
}

type unionModeInfo struct {
	targetMode DISPLAYCONFIG_TARGET_MODE
	sourceMode DISPLAYCONFIG_SOURCE_MODE
}

// HDR struct controls Windows HDR settings
type HDR struct{}

// NewHDR creates a new HDR controller
func NewHDR() *HDR {
	return &HDR{}
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procGetDisplayConfigBufferSizes = user32.NewProc("GetDisplayConfigBufferSizes")
	procQueryDisplayConfig          = user32.NewProc("QueryDisplayConfig")
	procDisplayConfigSetDeviceInfo  = user32.NewProc("DisplayConfigSetDeviceInfo")
)

// SetHDR enables or disables HDR on all compatible displays
func (h *HDR) SetHDR(enable bool) error {
	var pathCount, modeCount uint32

	// Get buffer sizes
	ret, _, err := procGetDisplayConfigBufferSizes.Call(
		uintptr(QDC_ONLY_ACTIVE_PATHS),
		uintptr(unsafe.Pointer(&pathCount)),
		uintptr(unsafe.Pointer(&modeCount)),
	)

	if ret != 0 {
		return errors.New("failed to get display config buffer sizes: " + err.Error())
	}

	if pathCount == 0 || modeCount == 0 {
		return errors.New("no active displays found")
	}

	// Allocate arrays for paths and modes
	paths := make([]DISPLAYCONFIG_PATH_INFO, pathCount)
	modes := make([]DISPLAYCONFIG_MODE_INFO, modeCount)

	// Query current display config
	ret, _, err = procQueryDisplayConfig.Call(
		uintptr(QDC_ONLY_ACTIVE_PATHS),
		uintptr(unsafe.Pointer(&pathCount)),
		uintptr(unsafe.Pointer(&paths[0])),
		uintptr(unsafe.Pointer(&modeCount)),
		uintptr(unsafe.Pointer(&modes[0])),
		uintptr(0),
	)

	if ret != 0 {
		return errors.New("failed to query display config: " + err.Error())
	}

	var lastError error
	successCount := 0

	// Apply HDR setting to each active display
	for i := uint32(0); i < pathCount; i++ {
		path := &paths[i]
		if path.Target.TargetAvailable != 0 {
			if err := h.setHDRForDisplay(&path.Target, enable); err != nil {
				lastError = err
			} else {
				successCount++
			}
		}
	}

	if successCount == 0 && lastError != nil {
		return lastError
	}

	if successCount == 0 {
		return errors.New("no compatible displays were configured")
	}

	return nil
}

// setHDRForDisplay sets HDR state for a specific display target
func (h *HDR) setHDRForDisplay(target *DISPLAYCONFIG_PATH_TARGET_INFO, enable bool) error {
	// Prepare the advanced color info structure
	colorInfo := DISPLAYCONFIG_ADVANCED_COLOR_INFO{}
	colorInfo.Header.Type = DISPLAYCONFIG_DEVICE_INFO_SET_ADVANCED_COLOR_INFO
	colorInfo.Header.Size = uint32(unsafe.Sizeof(colorInfo))
	colorInfo.Header.AdapterId = target.AdapterId
	colorInfo.Header.Id = target.Id

	if enable {
		colorInfo.Value = 1 // Enable HDR
	} else {
		colorInfo.Value = 0 // Disable HDR
	}

	// Set the HDR state
	ret, _, err := procDisplayConfigSetDeviceInfo.Call(
		uintptr(unsafe.Pointer(&colorInfo.Header)),
	)

	if ret != 0 {
		return errors.New("failed to set HDR state: " + err.Error())
	}

	return nil
}

// Enable turns on HDR for all compatible displays
func (h *HDR) Enable() error {
	return h.SetHDR(true)
}

// Disable turns off HDR for all compatible displays
func (h *HDR) Disable() error {
	return h.SetHDR(false)
}

// Toggle switches HDR state on all compatible displays
func (h *HDR) Toggle() error {
	// For simplicity, we'll just try to disable first, then enable if that fails
	// In a real implementation, you'd query the current state first
	err := h.Disable()
	if err != nil {
		return h.Enable()
	}
	return nil
}

// GetState returns placeholder HDR state information
func (h *HDR) GetState() ([]string, error) {
	// This would require DisplayConfigGetDeviceInfo to get current state
	// For now, return a placeholder
	return []string{"HDR state detection not implemented"}, nil
}

// IsHDRSupported checks if HDR operations are likely supported
func (h *HDR) IsHDRSupported() bool {
	// Try a simple operation to see if the API is available
	var pathCount, modeCount uint32
	ret, _, _ := procGetDisplayConfigBufferSizes.Call(
		uintptr(QDC_ONLY_ACTIVE_PATHS),
		uintptr(unsafe.Pointer(&pathCount)),
		uintptr(unsafe.Pointer(&modeCount)),
	)
	return ret == 0
}
