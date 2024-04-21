package device

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

// type VolumeConsumer interface {
// 	OnVolume([]int)
// 	OnMute([]bool)
// }

func OpenAndDispatch(ctx context.Context, portName string /*consumer VolumeConsumer*/) error {
	port, err := serial.Open(portName, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return err
	}
	defer port.Close()

	reader := bufio.NewReader(port)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// intentionaly empty
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSuffix(line, "\r\n")
		fmt.Printf("Read %q\n", line)
	}
}

func ListNames() ([]string, error) {
	return serial.GetPortsList()
}

func TryReconnect(ctx context.Context, portName string) error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		err := OpenAndDispatch(ctx, portName)
		if err == nil {
			return nil
		}

		log.Println("cannot connect: sleeping 5s:", err)
		// TODO: handle errors that should be fatal

		ticker.Reset(time.Second * 5)
		select {
		case <-ctx.Done():
			// context terminated - should exit
			return nil

		case <-ticker.C:
			// ticker hit - try connect again
		}
	}
}
