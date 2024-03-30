package deej

import (
	"log"
	"strconv"
)

type MuteMap struct {
	targets map[int][]string
}

func (mm *MuteMap) Get(position int) ([]string, bool) {
	targets, exists := mm.targets[position]
	return targets, exists
}

// Load values from viper configuration, and make new instance of muteMap.
func muteMapFromConfigs(configValues map[string][]string) *MuteMap {
	targets := make(map[int][]string, len(configValues))

	for key, values := range configValues {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			log.Printf("Key %q, is not a valid integer: %s", key, err.Error())
			continue
		}
		targets[intKey] = values
	}

	return &MuteMap{
		targets: targets,
	}
}
