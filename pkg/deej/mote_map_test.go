package deej

import "testing"

func TestMuteMap(t *testing.T) {
	givenConfig := map[string][]string{
		"0": {"mic"},
		"1": {"master"},
	}
	muteMapFromConfigs(givenConfig)
}
