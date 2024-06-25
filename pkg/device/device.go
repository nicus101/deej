package device

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

var ErrConnectionTimeout = errors.New("line read timeouted")

var lastPortName string

type Connection struct {
	portNameChannel chan string
}

type VolumeConsumer interface {
	OnVolume([]int)
	OnMute([]bool)
}

// TODO connection busy
func (ConnectAD *Connection) ConnectAndDispatch(
	ctx context.Context,
	portName string,
	volumeConsumer VolumeConsumer,
) error {
	log.Println("Connecting to:", portName)

	port, err := serial.Open(portName, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return err
	}
	defer port.Close()

	if ConnectAD.portNameChannel == nil {
		ConnectAD.portNameChannel = make(chan string, 1)
	}
	lastPortName = portName

	timerTimeout := time.Second * 5
	timerHit := false
	timer := time.AfterFunc(timerTimeout, func() {
		timerHit = true
		port.Close()
	})

	reader := bufio.NewReader(port)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case newPortName := <-ConnectAD.portNameChannel:
			log.Println("Changing port to:", newPortName)
			port.Close()
			return ConnectAD.ConnectAndDispatch(ctx, newPortName, volumeConsumer)

		default:
			// intentionaly empty
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if timerHit {
				err = ErrConnectionTimeout
			}
			// maybe add ErrConnectionClose na wypadek "device unplugged"?
			return err
		}
		timer.Reset(timerTimeout)

		line = strings.TrimSuffix(line, "\r\n")
		fmt.Printf("Read %q\n", line)
		parseAndDispatch(line, volumeConsumer)
	}
}

// zrób metodę na wskaźniku Connection o następującej geometri DevicePortSet(deviceName string)
func (ConnectAD *Connection) DevicePortSet(deviceName string) {
	if ConnectAD == nil {
		return
	}

	ConnectAD.portNameChannel <- deviceName
}

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
	if deviceName := lastPortName; deviceName != "" {
		deviceFound = append(deviceFound, deviceName)
	}

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
