package ui

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/fstanis/screenresolution"
	"github.com/gen2brain/beeep"
	"github.com/gonutz/wui/v2"
	"github.com/omriharel/deej/pkg/device"
	"golang.org/x/exp/maps"
)

// TODO: make abstraction

// type appComboEntry struct {
// 	name string

// 	// przechowywać infotmacje że aplikacja jest:
// 	// - zaznaczona
// 	// - zapisana
// 	// - aktualnie nie uruchomiona
// }

// type appComboBox struct {
// 	*wui.ComboBox

// 	apps []appComboBox
// }

const selectedPrefix = "✔ "

func setComboAppList(
	comboBox *wui.ComboBox,
	appList []string,
	configuredApps []string,
) {
	log.Println(
		"Refreshing appList:", strings.Join(appList, ", "),
		"\n\tWith configured:", strings.Join(configuredApps, ", "),
	)

	// połączenie aktualnych aplikacji, razem z skonfigurowanymi aplikacjami
	// bez powtórzeń, z priorytetem dla skonfigurowanych appek.
	appMap := make(map[string]string, len(appList)+len(configuredApps))
	for _, appName := range appList {
		appMap[appName] = appName
	}
	for _, appName := range configuredApps {
		appMap[appName] = selectedPrefix + appName
	}
	appKeys := maps.Keys(appMap)
	sort.Strings(appKeys)

	// we want virtual first entry that shows on combo box closed
	// and inform user what apps are currently selected for
	// controll channel coresponding with this combo box
	var headerString string
	switch l := len(configuredApps); {

	case l == 0:
		headerString = "= Nothing ="

	case l < 4:
		headerString = strings.Join(configuredApps, " & ")

	default:
		headerString = fmt.Sprintf("%d app selected", len(configuredApps))
	}

	// using deduplicated appKeys to make label listy, for user showing
	items := make([]string, 0, len(appKeys)+1)
	items = append(items, headerString)
	for _, appKey := range appKeys {
		items = append(items, appMap[appKey])
	}

	comboBox.SetItems(items)
	comboBox.SetSelectedIndex(0)
}

func makeAppSelectFunc(
	_ int,
	_ ChannelAppsSetter,
	appList []string,
	comboBox *wui.ComboBox,
) func(int) {
	return func(i int) {
		log.Printf("ComboBox %p selected %q\n", comboBox, appList[i])

		if strings.HasPrefix(appList[i], selectedPrefix) {
			appList[i] = appList[i][len(selectedPrefix):]
		} else {
			appList[i] = selectedPrefix + appList[i]
		}

		selectedApps := make([]string, 0, len(appList))
		for _, appName := range appList {
			if strings.HasPrefix(appName, selectedPrefix) {
				selectedApps = append(selectedApps, appName)
			}
		}

		selectedLabel := "= nothing ="
		if len(selectedApps) > 3 {
			selectedLabel = fmt.Sprintf("%d app selected", len(selectedApps))
		} else if len(selectedApps) > 0 {
			selectedLabel = strings.Join(selectedApps, " & ")
		}

		appList[0] = selectedLabel
		comboBox.SetItems(appList)
		comboBox.SetSelectedIndex(0)
	}
}

type DevicePortSetter interface {
	DevicePortSet(deviceName string)
}

type ProgramLister interface {
	ProgramList() ([]string, error)
}

type SettingsWriteCanceler interface {
	// zapisuje aktualny konfig na dysk
	Write() error

	// kanceluje zmiany i wczytuje stare nastawy
	Cancel() error
}

type ChannelAppsSetter interface {
	ChannelAppsSet(chanId int, apps []string)
	ChannelAppGet(chanId int) []string
}

func ShowUI(
	_ []string,
	devicePortSetter DevicePortSetter,
	programLister ProgramLister,
	settingsWriteCanceler SettingsWriteCanceler,
	channelAppsSetter ChannelAppsSetter,
) {
	appList, err := programLister.ProgramList()
	if err != nil {
		log.Println("Cannot get program list:", err)
	}

	appList = append(appList, "deej.unmapped", "deej.current")

	configWindowFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -11,
	})

	configChanged := false

	resolution := screenresolution.GetPrimary()
	configWindow := wui.NewWindow()

	configWindow.SetFont(configWindowFont)
	configWindow.SetInnerY(46)
	configWindow.SetInnerSize(315, 340)
	configWindow.SetTitle("Mixer Deck Configurator")
	configWindow.SetHasMaxButton(false)
	configWindow.SetResizable(false)
	configWindow.SetOnCanClose(func() bool {
		return askClose(configChanged)
	})

	panel3Font, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -11,
	})

	panel3 := wui.NewPanel()
	panel3.SetFont(panel3Font)
	panel3.SetBounds(7, 3, 300, 55)
	panel3.SetBorderStyle(wui.PanelBorderSingleLine)
	configWindow.Add(panel3)

	panel4Font, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -11,
	})

	panel4 := wui.NewPanel()
	panel4.SetFont(panel4Font)
	panel4.SetBounds(7, 61, 300, 246)
	panel4.SetBorderStyle(wui.PanelBorderSingleLine)
	configWindow.Add(panel4)

	devices, err := device.ListAllNames()
	if err != nil {
		log.Println("Cannot list ports:", err)
	}
	devices = append([]string{"Chose port"}, devices...)

	comPortBox := wui.NewComboBox()
	comPortBox.SetBounds(100, 10, 100, 20)
	comPortBox.SetItems(devices)
	comPortBox.SetSelectedIndex(0)
	comPortLastSelected := 0
	comPortBox.SetOnChange(func(selectedPort int) {
		if selectedPort != comPortLastSelected {
			configChanged = true
		}
	})
	configWindow.Add(comPortBox)

	detectButton := wui.NewButton()
	detectButton.SetBounds(100, 35, 100, 20)
	detectButton.SetText("Detect port")

	configWindow.Add(detectButton)

	//================================================================ Start of pots combo boxes ================================================================
	comboBoxes := make([]*wui.ComboBox, 6)
	for i := 0; i < 6; i++ {
		y := 72 + i*35
		comboBoxes[i] = addComboBox(configWindow, i, 100, y, appList, channelAppsSetter)
	}

	//================================================================= Labels ============================================================================
	addLabel(configWindow, 15, 70, 80, 24, "Volume 1:")
	addLabel(configWindow, 15, 105, 80, 24, "Volume 2:")
	addLabel(configWindow, 15, 140, 80, 24, "Volume 3:")
	addLabel(configWindow, 15, 175, 80, 24, "Volume 4:")
	addLabel(configWindow, 15, 210, 80, 24, "Volume 5:")
	addLabel(configWindow, 15, 245, 80, 24, "Volume 6:")

	rescanBtn := wui.NewButton()
	rescanBtn.SetBounds(185, 278, 100, 20)
	rescanBtn.SetText("Rescan")
	addLabel(configWindow, 25, 275, 150, 24, "Don't see your app on list HIT =>")

	configWindow.Add(rescanBtn)

	cancelBtn := wui.NewButton()
	cancelBtn.SetBounds(7, 310, 85, 25)
	cancelBtn.SetText("Close")
	cancelBtn.SetOnClick(func() {
		configWindow.Close()
	})
	configWindow.Add(cancelBtn)

	detectedDeviceName := ""
	saveBtn := wui.NewButton()
	saveBtn.SetBounds(222, 310, 85, 25)
	saveBtn.SetText("Save")
	saveBtn.SetOnClick(func() {
		log.Print("Device ", detectedDeviceName)
		if devicePortSetter == nil {
			return
		}
		if detectedDeviceName == "" {
			return
		}
		devicePortSetter.DevicePortSet(detectedDeviceName)
	})
	configWindow.Add(saveBtn)

	addLabel(configWindow, 15, 10, 83, 27, "Chose port:")
	isConnectedLabel := addLabel(configWindow, 207, 10, 90, 10, "Click detect port")

	applyBtn := wui.NewButton()
	applyBtn.SetBounds(114, 310, 85, 25)
	applyBtn.SetText("Test")
	applyBtn.SetOnClick(func() {
		log.Print("Device ", detectedDeviceName)
		if devicePortSetter == nil {
			return
		}
		if detectedDeviceName == "" {
			return
		}
		configChanged = true
		devicePortSetter.DevicePortSet(detectedDeviceName)
	})
	configWindow.Add(applyBtn)

	rescanBtn.SetOnClick(func() {
		scannedApps, err := programLister.ProgramList()
		if err != nil {
			log.Println("Can't get apps list", err)
			return
		}
		if len(scannedApps) == 0 {
			log.Println("No apps detected.")
			return
		}

		for chanId, comboBox := range comboBoxes {
			configuredApps := channelAppsSetter.ChannelAppGet(chanId)
			setComboAppList(comboBox, scannedApps, configuredApps)
		}

		configWindow.Repaint()
	})

	detectButton.SetOnClick(func() {
		detectedDevices, err := device.ListNames()
		if err != nil {
			log.Println("Can't get devices list", err)
			return
		}
		if len(detectedDevices) == 0 {
			log.Println("No devices detected.")
			return
		}
		if len(detectedDevices) != 1 {
			responseString := fmt.Sprint("More than one device detected try manually selecting one of those ports:", detectedDevices)
			err := beeep.Notify("Warning!", responseString, "")
			if err != nil {
				log.Println("Error", err)
			}
			return
		}
		comPortBox.SetItems(detectedDevices)
		comPortBox.SetSelectedIndex(0)
		isConnectedLabel.SetText("Detected port:")
		addLabel(configWindow, 207, 30, 90, 10, detectedDevices[0])
		detectedDeviceName = detectedDevices[0]
		configWindow.Repaint()

	})

	configWindow.SetPosition(
		resolution.Width/2-configWindow.Width()/2,
		resolution.Height/2-configWindow.Height()/2,
	)
	configWindow.Show()

	configWindow.Repaint()
}

func addComboBox(
	wnd *wui.Window,
	i, x, y int,
	appList []string,
	channelAppsSetter ChannelAppsSetter,
) *wui.ComboBox {
	appList = append([]string{"= nothing ="}, appList...)
	comboBox := wui.NewComboBox()
	comboBox.SetBounds(x, y, 200, 24)
	comboBox.SetItems(appList)
	comboBox.SetSelectedIndex(0)
	comboBox.SetOnChange(makeAppSelectFunc(
		i, channelAppsSetter,
		appList,
		comboBox,
	))
	wnd.Add(comboBox)
	return comboBox
}

func addLabel(
	wnd *wui.Window,
	x, y, sx, sy int,
	labelText string,
) *wui.Label {
	label := wui.NewLabel()
	label.SetBounds(x, y, sx, sy)
	label.SetText(labelText)
	wnd.Add(label)
	return label
}

func askClose(configChanged bool) bool {

	if !configChanged {
		return true
	}

	askWindowFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -11,
	})

	resolution := screenresolution.GetPrimary()
	askWindow := wui.NewWindow()
	askWindow.SetFont(askWindowFont)
	askWindow.SetInnerSize(261, 101)
	askWindow.SetTitle("Configuration changed")
	askWindow.SetPosition(
		resolution.Width/2-askWindow.Width()/2,
		resolution.Height/2-askWindow.Height()/2,
	)

	shouldClose := false

	askYes := wui.NewButton()
	askYes.SetBounds(10, 68, 85, 25)
	askYes.SetText("Yes")
	askYes.SetOnClick(func() {
		askWindow.Close()
		shouldClose = true
	})
	askWindow.Add(askYes)

	askNo := wui.NewButton()
	askNo.SetBounds(168, 68, 85, 25)
	askNo.SetText("No")
	askNo.SetOnClick(func() {
		askWindow.Close()
	})
	askWindow.Add(askNo)

	askLabel := wui.NewLabel()
	askLabel.SetHorizontalAnchor(wui.AnchorCenter)
	askLabel.SetBounds(4, 5, 254, 53)
	askLabel.SetText("Do you wish to exit without saving?")
	askLabel.SetAlignment(wui.AlignCenter)
	askWindow.Add(askLabel)

	askWindow.ShowModal()
	askWindow.Repaint()
	return shouldClose
}
