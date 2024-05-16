package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gonutz/wui/v2"
)

func makeAppSelectFunc(appList []string, comboBox *wui.ComboBox) func(int) {
	selectedPrefix := "✔ "
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

func ShowUI(appList []string, devicePortSetter DevicePortSetter) {
	appList = append(appList, "deej.unmapped", "deej.current")

	configWindowFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -11,
	})

	configWindow := wui.NewWindow()
	configWindow.SetFont(configWindowFont)
	configWindow.SetInnerY(46)
	configWindow.SetInnerSize(315, 317)
	configWindow.SetTitle("Mixer Deck Configurator")
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
	panel4.SetBounds(7, 61, 300, 216)
	panel4.SetBorderStyle(wui.PanelBorderSingleLine)
	configWindow.Add(panel4)

	comPortBox := wui.NewComboBox()
	comPortBox.SetBounds(100, 10, 100, 20)
	comPortBox.SetItems([]string{"Wybierz port"})
	comPortBox.SetSelectedIndex(0)
	configWindow.Add(comPortBox)

	//================================================================ Start of pots combo boxes ================================================================
	addComboBox(configWindow, 100, 65, appList)
	addComboBox(configWindow, 100, 100, appList)
	addComboBox(configWindow, 100, 135, appList)
	addComboBox(configWindow, 100, 170, appList)
	addComboBox(configWindow, 100, 205, appList)
	addComboBox(configWindow, 100, 245, appList)

	//================================================================= Labels ============================================================================
	addLabel(configWindow, 15, 70, 80, 24, "Potencjometr 1:")
	addLabel(configWindow, 15, 105, 80, 24, "Potencjometr 2:")
	addLabel(configWindow, 15, 140, 80, 24, "Potencjometr 3:")
	addLabel(configWindow, 15, 175, 80, 24, "Potencjometr 4:")
	addLabel(configWindow, 15, 210, 80, 24, "Potencjometr 5:")
	addLabel(configWindow, 15, 245, 80, 24, "Potencjometr 6:")

	cancelBtn := wui.NewButton()
	cancelBtn.SetBounds(7, 282, 85, 25)
	cancelBtn.SetText("Anuluj")
	configWindow.Add(cancelBtn)

	saveBtn := wui.NewButton()
	saveBtn.SetBounds(222, 282, 85, 25)
	saveBtn.SetText("Zapisz")
	configWindow.Add(saveBtn)

	addLabel(configWindow, 15, 10, 83, 27, "Wybierz port:")
	addLabel(configWindow, 207, 10, 90, 40, "Połączony?")

	applyBtn := wui.NewButton()
	applyBtn.SetBounds(114, 282, 85, 25)
	applyBtn.SetText("Zastosuj")
	configWindow.Add(applyBtn)

	configWindow.Show()
	configWindow.Repaint()
}

func addComboBox(
	wnd *wui.Window,
	x, y int,
	appList []string,
) {
	appList = append([]string{"= nothing ="}, appList...)
	comboBox := wui.NewComboBox()
	comboBox.SetBounds(x, y, 200, 24)
	comboBox.SetItems(appList)
	comboBox.SetSelectedIndex(0)
	comboBox.SetOnChange(makeAppSelectFunc(
		appList,
		comboBox,
	))
	wnd.Add(comboBox)
}

func addLabel(
	wnd *wui.Window,
	x, y, sx, sy int,
	labelText string,
) {
	label := wui.NewLabel()
	label.SetBounds(x, y, sx, sy)
	label.SetText(labelText)
	wnd.Add(label)
}
