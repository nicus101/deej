package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/omriharel/deej/pkg/device"
)

type fakeMikser struct{}

var _ device.VolumeConsumer = fakeMikser{}

func (fakeMikser) OnVolume(volumes []int) {
	var builder strings.Builder
	fmt.Fprint(&builder, "OnVolume:")

	for i, volume := range volumes {
		fmt.Fprint(&builder, " ", volume)
		if i != 0 {
			fmt.Fprint(&builder, ",")
		}
	}

	log.Println(builder.String())
}

func (fakeMikser) OnMute(mutes []bool) {
	var builder strings.Builder
	fmt.Fprint(&builder, "OnMute:")

	for i, mute := range mutes {
		fmt.Fprint(&builder, " ", mute)
		if i != 0 {
			fmt.Fprint(&builder, ",")
		}
	}

	log.Println(builder.String())

}
