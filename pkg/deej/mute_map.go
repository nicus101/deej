package deej

import "fmt"

type MuteMap struct {
}

// Consume bool slice, made by serial.go, and propagate to correct consumers.
func (muteMap *MuteMap) Mute(values []bool) {

}

// Load values from viper configuration, and make new instance of muteMap.
func muteMapFromConfigs(configValues map[string][]string) *MuteMap {
	fmt.Println(configValues)
	return nil
}
