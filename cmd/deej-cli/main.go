package main

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/moutend/go-wca"
	"github.com/omriharel/deej/pkg/session"
)

func main() {
	// before we can do anything with windows subsystems - we need OLE
	closeFn, err := session.OleConnectAndCleanUp()
	if err != nil {
		log.Fatalln("Something went wrong when connecting to OS:", err)
	}
	defer closeFn()

	// device enumerator
	mmDeviceEnumerator, err := session.GetDeviceEnumaerator()
	if err != nil {
		log.Fatalln("Something went wrong when connecting to OS:", err)
	}
	defer mmDeviceEnumerator.Release()

	var mmOutDevice *wca.IMMDevice
	err = mmDeviceEnumerator.GetDefaultAudioEndpoint(
		wca.ERender,
		wca.EConsole,
		&mmOutDevice,
	)
	if err != nil {
		log.Fatalln("output device failed:", err)
	}
	defer func() {
		if mmOutDevice != nil {
			mmOutDevice.Release()
		}
	}()

	var mmInDevice *wca.IMMDevice
	err = mmDeviceEnumerator.GetDefaultAudioEndpoint(
		wca.ECapture,
		wca.EConsole,
		&mmInDevice,
	)
	if err != nil {
		log.Fatalln("input device failed:", err)
	}
	defer func() {
		if mmInDevice != nil {
			mmInDevice.Release()
		}
	}()

	// TODO: something something notification

	// get list of devices
	var deviceCollection *wca.IMMDeviceCollection
	err = mmDeviceEnumerator.EnumAudioEndpoints(
		wca.EAll,
		wca.DEVICE_STATE_ACTIVE,
		&deviceCollection,
	)
	if err != nil {
		log.Fatalln("enumerate active audio endpoints: %w", err)
	}

	// check how many devices there are
	var deviceCount uint32
	err = deviceCollection.GetCount(&deviceCount)
	if err != nil {
		log.Fatalln("get device count from device collection: %w", err)
	}

	// for each device:
	for deviceIdx := uint32(0); deviceIdx < deviceCount; deviceIdx++ {
		err = scanSomething(deviceCollection, deviceIdx)
		if err != nil {
			log.Fatalln("something went wrong:", err)
		}
	}
}

func scanSomething(deviceCollection *wca.IMMDeviceCollection, deviceIdx uint32) (err error) {

	// get its IMMDevice instance
	var endpoint *wca.IMMDevice
	err = deviceCollection.Item(deviceIdx, &endpoint)
	if err != nil {
		return fmt.Errorf("get device %d from device collection: %w", deviceIdx, err)
	}
	defer endpoint.Release()

	// get its IMMEndpoint instance to figure out if it's an output device (and we need to enumerate its process sessions later)
	dispatch, err := endpoint.QueryInterface(wca.IID_IMMEndpoint)
	if err != nil {
		return fmt.Errorf("query device %d IMMEndpoint: %w", deviceIdx, err)
	}

	// get the device's property store
	var propertyStore *wca.IPropertyStore
	err = endpoint.OpenPropertyStore(wca.STGM_READ, &propertyStore)
	if err != nil {
		return fmt.Errorf("open endpoint %d property store: %w", deviceIdx, err)
	}
	defer propertyStore.Release()

	// query the property store for the device's description and friendly name
	value := new(wca.PROPVARIANT)
	err = propertyStore.GetValue(&wca.PKEY_Device_DeviceDesc, value)
	if err != nil {
		return fmt.Errorf("get device %d description: %w", deviceIdx, err)
	}

	// device description i.e. "Headphones"
	endpointDescription := strings.ToLower(value.String())
	err = propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value)
	if err != nil {
		return fmt.Errorf("get device %d friendly name: %w", deviceIdx, err)
	}

	// device friendly name i.e. "Headphones (Realtek Audio)"
	endpointFriendlyName := value.String()

	// receive a useful object instead of our dispatch
	endpointType := (*wca.IMMEndpoint)(unsafe.Pointer(dispatch))
	defer endpointType.Release()

	var dataFlow uint32
	err = endpointType.GetDataFlow(&dataFlow)
	if err != nil {
		return fmt.Errorf("get device %d data flow: %w", deviceIdx, err)
	}

	log.Println("Enumerated device info ",
		"deviceIdx ", deviceIdx,
		"deviceDescription ", endpointDescription,
		"deviceFriendlyName ", endpointFriendlyName,
		"dataFlow ", dataFlow)

	// if the device is an output device, enumerate and add its per-process audio sessions
	if dataFlow == wca.ERender {
		err = enumerateOrSomething(endpoint)
		if err != nil {
			return fmt.Errorf("enumerate and add device %d process sessions: %w", deviceIdx, err)
		}
	}

	// // for all devices (both input and output), add a named "master" session that can be addressed
	// // by using the device's friendly name (as appears when the user left-clicks the speaker icon in the tray)
	// newSession, err := sf.getMasterSession(endpoint,
	// 	endpointFriendlyName,
	// 	fmt.Sprintf(deviceSessionFormat, endpointDescription))

	// if err != nil {
	// 	sf.logger.Warnw("Failed to get master session for device",
	// 		"deviceIdx", deviceIdx,
	// 		"error", err)

	// 	return fmt.Errorf("get device %d master session: %w", deviceIdx, err)
	// }
	return nil
}

func enumerateOrSomething(endpoint *wca.IMMDevice) (err error) {

	log.Println("Enumerating and adding process sessions for audio output device",
		"deviceFriendlyName", " = UwU = ")

	// query the given IMMDevice's IAudioSessionManager2 interface
	var audioSessionManager2 *wca.IAudioSessionManager2
	err = endpoint.Activate(
		wca.IID_IAudioSessionManager2,
		wca.CLSCTX_ALL,
		nil,
		&audioSessionManager2,
	)
	if err != nil {
		return fmt.Errorf("activate endpoint: %w", err)
	}
	defer audioSessionManager2.Release()

	// get its IAudioSessionEnumerator
	var sessionEnumerator *wca.IAudioSessionEnumerator
	err = audioSessionManager2.GetSessionEnumerator(&sessionEnumerator)
	if err != nil {
		return err
	}
	defer sessionEnumerator.Release()

	// check how many audio sessions there are
	var sessionCount int
	err = sessionEnumerator.GetCount(&sessionCount)
	if err != nil {
		return fmt.Errorf("get session count: %w", err)
	}

	log.Println("Got session count from session enumerator", "count", sessionCount)

	// for each session:
	var audioSessionControl *wca.IAudioSessionControl
	for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {

		err = sessionEnumerator.GetSession(sessionIdx, &audioSessionControl)
		if err != nil {
			return fmt.Errorf("get session %d from enumerator: %w", sessionIdx, err)
		}

		channel, err := session.MakeChannel(audioSessionControl, nil)
		audioSessionControl.Release()
		if err != nil {
			return fmt.Errorf("%d: %w", sessionIdx, err)
		}

		log.Printf("Found session %q", channel.Executable())
		channel.Release()
	}
	return nil
}
