package ui

import (
	"bytes"
	"fmt"
	"log"
	"slices"

	"github.com/fstanis/screenresolution"
	"github.com/gen2brain/beeep"
	"github.com/gonutz/wui/v2"
	"github.com/omriharel/deej/pkg/deej/icon"
	"github.com/omriharel/deej/pkg/device"
)

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
	mainIcon, _ := wui.NewIconFromReader(bytes.NewBuffer(icon.DeejLogo))

	configWindow.SetFont(configWindowFont)
	configWindow.SetInnerY(46)
	configWindow.SetInnerSize(315, 340)
	configWindow.SetTitle("MD Configurator")
	configWindow.SetIcon(mainIcon)
	configWindow.SetHasMaxButton(false)
	configWindow.SetResizable(false)

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

	detectedDeviceName := ""
	comPortBox.SetOnChange(func(selectedPort int) {
		if selectedPort != comPortLastSelected {
			configChanged = true
		}
		detectedDeviceName = comPortBox.Items()[selectedPort]
	})
	configWindow.Add(comPortBox)

	detectButton := wui.NewButton()
	detectButton.SetBounds(100, 35, 100, 20)
	detectButton.SetText("Detect port")

	configWindow.Add(detectButton)

	//================================================================ Start of pots combo boxes ================================================================
	comboBoxes := make([]*appComboBox, 6)
	for i := 0; i < 6; i++ {
		y := 72 + i*35
		comboBoxes[i] = newComboBox(configWindow, 100, y, appList)
		configuredApps := channelAppsSetter.ChannelAppGet(i)
		comboBoxes[i].populateAppConfigStatus(configuredApps)
	}
	configWindow.SetOnCanClose(func() bool {
		return askClose(configChanged, comboBoxes)
	})

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

	saveBtn := wui.NewButton()
	saveBtn.SetBounds(222, 310, 85, 25)
	saveBtn.SetText("Save")
	saveBtn.SetOnClick(func() {
		log.Print("Device ", detectedDeviceName)
		if devicePortSetter != nil && detectedDeviceName != "" {
			devicePortSetter.DevicePortSet(detectedDeviceName)
		}

		if channelAppsSetter != nil {
			for i, comboBox := range comboBoxes {
				channelAppsSetter.ChannelAppsSet(i, comboBox.toSaveAppNames())
			}
		}

		if settingsWriteCanceler != nil {
			settingsWriteCanceler.Write()
		}
		configChanged = false
		for i, comboBox := range comboBoxes {
			configuredApps := channelAppsSetter.ChannelAppGet(i)
			comboBox.populateAppConfigStatusAfterSave(configuredApps)
		}
	})
	configWindow.Add(saveBtn)

	addLabel(configWindow, 15, 10, 83, 27, "Chose port:")
	isConnectedLabel := addLabel(configWindow, 207, 10, 90, 10, "Click detect port")
	detectPorts(comPortBox, isConnectedLabel, configWindow, &detectedDeviceName)

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

		for i, comboBox := range comboBoxes {
			channelAppsSetter.ChannelAppsSet(i, comboBox.toSaveAppNames())
		}
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
			comboBox.populateAppList(scannedApps)

			configuredApps := channelAppsSetter.ChannelAppGet(chanId)
			comboBox.populateAppConfigStatus(configuredApps)
		}

		configWindow.Repaint()
	})

	detectButton.SetOnClick(func() {
		detectPorts(comPortBox, isConnectedLabel, configWindow, &detectedDeviceName)
	})

	configWindow.SetPosition(
		resolution.Width/2-configWindow.Width()/2,
		resolution.Height/2-configWindow.Height()/2,
	)
	configWindow.Show()

	configWindow.Repaint()
}

func detectPorts(comPortBox *wui.ComboBox, isConnectedLabel *wui.Label, configWindow *wui.Window, detectedDeviceName *string) {
	isConnectedLabel.SetText("Detecting...")

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
		//return
	}
	detectedDevices = slices.Compact(detectedDevices)
	comPortBox.SetItems(detectedDevices)
	comPortBox.SetSelectedIndex(0)
	isConnectedLabel.SetText("Detected port:")
	addLabel(configWindow, 207, 30, 90, 10, detectedDevices[0])
	*detectedDeviceName = detectedDevices[0]
	configWindow.Repaint()
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

func ConfigInfo() {
	configInfoWindowFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -14,
		Bold:   true,
	})

	resolution := screenresolution.GetPrimary()

	configInfoWindow := wui.NewWindow()
	configInfoWindow.SetFont(configInfoWindowFont)
	configInfoWindow.SetInnerSize(400, 130)
	configInfoWindow.SetTitle("Mixer device not detected!")
	configInfoWindow.SetHasMinButton(false)
	configInfoWindow.SetHasMaxButton(false)
	configInfoWindow.SetResizable(false)
	configInfoWindow.SetPosition(
		resolution.Width/2-configInfoWindow.Width()/2,
		resolution.Height/2-configInfoWindow.Height()/2,
	)

	confiInfoLabel1 := wui.NewLabel()
	confiInfoLabel1.SetAnchors(wui.AnchorCenter, wui.AnchorCenter)
	confiInfoLabel1.SetBounds(10, 10, 380, 20)
	confiInfoLabel1.SetText("Can't detect mixer device,")
	confiInfoLabel1.SetAlignment(wui.AlignCenter)
	configInfoWindow.Add(confiInfoLabel1)

	confiInfoLabel2 := wui.NewLabel()
	confiInfoLabel2.SetAnchors(wui.AnchorCenter, wui.AnchorCenter)
	confiInfoLabel2.SetBounds(10, 30, 380, 20)
	confiInfoLabel2.SetText("make sure it is connected to PC,")
	confiInfoLabel2.SetAlignment(wui.AlignCenter)
	configInfoWindow.Add(confiInfoLabel2)

	confiInfoLabel3 := wui.NewLabel()
	confiInfoLabel3.SetAnchors(wui.AnchorCenter, wui.AnchorCenter)
	confiInfoLabel3.SetBounds(10, 51, 380, 20)
	confiInfoLabel3.SetText("or find Mixer-deej icon on tray and open configuration editor")
	confiInfoLabel3.SetAlignment(wui.AlignCenter)
	configInfoWindow.Add(confiInfoLabel3)

	confiInfoLabel4 := wui.NewLabel()
	confiInfoLabel4.SetAnchors(wui.AnchorCenter, wui.AnchorCenter)
	confiInfoLabel4.SetBounds(10, 70, 380, 20)
	confiInfoLabel4.SetText("Then push detect device button.")
	confiInfoLabel4.SetAlignment(wui.AlignCenter)
	configInfoWindow.Add(confiInfoLabel4)

	okBtn := wui.NewButton()
	okBtn.SetVerticalAnchor(wui.AnchorCenter)
	okBtn.SetBounds(160, 95, 80, 30)
	okBtn.SetText("OK")
	okBtn.SetOnClick(func() {
		configInfoWindow.Close()
	})
	configInfoWindow.Add(okBtn)

	configInfoWindow.ShowModal()
	configInfoWindow.Repaint()
}

func askClose(configChanged bool, comboBoxes []*appComboBox) bool {
	for _, comboBox := range comboBoxes {
		if comboBox.haveUserChanges() {
			configChanged = true
		}
	}

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
