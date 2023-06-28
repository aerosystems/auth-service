package validators

import (
	"errors"
	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
)

func ValidateDeviceType(deviceType string) error {
	if !helpers.Contains([]string{"web", "android", "ios"}, deviceType) {
		return errors.New("invalid device type")
	}
	return nil
}
