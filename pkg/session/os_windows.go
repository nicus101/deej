package session

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca"
)

// TODO: make private
func OleConnectAndCleanUp() (func(), error) {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err == nil {
		return ole.CoUninitialize, nil
	}

	var oleError *ole.OleError
	// check for E_FALSE - it means reduntant initialisation - no error
	if errors.As(err, &oleError) && oleError.Code() == 1 {
		log.Println("CoInitializeEx: E_FALSE - reduntand but not fatal")
		return func() {}, nil
	}

	return nil, fmt.Errorf("ole initialisation: %w", err)
}

func GetDeviceEnumaerator() (*wca.IMMDeviceEnumerator, error) {
	var mmDeviceEnumerator *wca.IMMDeviceEnumerator
	err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&mmDeviceEnumerator,
	)
	return mmDeviceEnumerator, err
}
