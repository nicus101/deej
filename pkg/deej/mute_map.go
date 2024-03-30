package deej

import "fmt"

type MuteMap struct {
}

func (mm *MuteMap) Get(position int) ([]string, bool) {

	return nil, false
}

// Load values from viper configuration, and make new instance of muteMap.
func muteMapFromConfigs(configValues map[string][]string) *MuteMap {
	fmt.Println(configValues)
	return nil
}
