package device

// import (
// 	"log"

// 	"github.com/omriharel/deej/pkg/device"
// )

// type Connection struct {
// 	appList []string
// 	comList []string
// }

// func getPortList() (portList []string) {
// 	portList, err := device.ListNames()
// 	if err != nil {
// 		log.Printf("Unable to get port list\n %s", portList)
// 	}
// 	return portList
// }

// func findPort(portList []string) {
// 	for i, portName := range portList {
// 		device.OpenAndDispatch()
// 	}
// }
