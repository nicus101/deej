package device

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"go.bug.st/serial"
)

func OpenAndDispatch(ctx context.Context, portName string) error {
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
