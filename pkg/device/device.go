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

func ListAllNames() ([]string, error) {
	return serial.GetPortsList()
}

func ListNames() ([]string, error) {
	portList, err := ListAllNames()

	if err != nil {
		return nil, fmt.Errorf("can't get port list: %w", err)
	}

	var deviceFound []string
	for _, portName := range portList {
		log.Println("Detecting device on:", portName)

		port, err := serial.Open(portName, &serial.Mode{
			BaudRate: 9600,
		})
		if err != nil {
			log.Printf("can't open port: %s", err)
			continue
		}
		defer port.Close()

		go func() {
			time.Sleep(5 * time.Second)
			port.Close()
		}()

		reader := bufio.NewReader(port)

		for i := 0; i < 3; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("can't read port: %s", err)
				continue
			}
			fmt.Printf("Read %q\n", line)

			if !strings.HasPrefix(line, "but|") {
				continue
			}
			if !strings.HasSuffix(line, "\r\n") {
				continue
			}

			log.Println("device found at:", portName)
			deviceFound = append(deviceFound, portName)
		}
	}

	return deviceFound, nil
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
