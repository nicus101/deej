package device

import (
	"log"
	"strconv"
	"strings"
)

func parseAndDispatch(line string, volumeConsumer VolumeConsumer) {
	isMute := false
	if strings.HasPrefix(line, "but|") {
		isMute = true
	}
	line = strings.TrimPrefix(line, "but|")

	values := strings.Split(line, "|")

	var volumes []int
	var mutes []bool

	for _, value := range values {
		if isMute {
			mute, err := strconv.ParseBool(value)
			if err != nil {
				log.Print(err)
				return
			}
			mutes = append(mutes, mute)
		} else {

			volume, err := strconv.Atoi(value)
			if err != nil {
				log.Print(err)
				return
			}
			volumes = append(volumes, volume)

		}
	}

	if volumeConsumer == nil {
		return
	}
	if isMute {
		volumeConsumer.OnMute(mutes)
	} else {
		volumeConsumer.OnVolume(volumes)
	}
	// jeżeli volumeConsumer nil, zreturnuj

	// jeżeli isMute wywołaj onMute

	// else wywołaj onVolume

	// zadanie na 6 - co zrobić z tym i, i wypisanie na ekran nie będzie szósteczką ;)

}
