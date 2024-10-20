package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/omriharel/deej/pkg/device"
)

func main() {
	if len(os.Args) != 2 {
		printPortNames()
		return
	}

	portName := os.Args[1]
	log.Println("Opening device at:", portName)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &device.Connection{}

	err := connection.ConnectAndDispatch(ctx, portName, fakeMikser{})
	if err != nil {
		log.Fatalln("Cannot open device: ", err)
	}
}

func printPortNames() {
	portNames, err := device.ListNames()
	if err != nil {
		log.Fatalln("Cannot list devices: ", err)
	}

	fmt.Println("Avaliable ports:\n\t", strings.Join(portNames, "\n\t"))
}
