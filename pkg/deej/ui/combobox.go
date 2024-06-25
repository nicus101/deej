package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gonutz/wui/v2"
	"golang.org/x/exp/maps"
)

const selectedPrefix = "✔ "

type appComboStatus struct {
	isConfigured bool
	isSelected   bool
}

type appComboBox struct {
	*wui.ComboBox
	appMap map[string]appComboStatus
}

func newComboBox(
	wnd *wui.Window,
	x, y int,
	appNames []string,
) *appComboBox {
	cb := &appComboBox{
		ComboBox: wui.NewComboBox(),
		appMap:   make(map[string]appComboStatus),
	}
	cb.SetBounds(x, y, 200, 24)
	cb.populateAppList(appNames)
	cb.updateAppList()
	cb.SetOnChange(cb.onChange)

	wnd.Add(cb)
	return cb
}

func (cb *appComboBox) toSaveAppNames() []string {
	appNames := make([]string, 0, len(cb.appMap))
	for appName, appStatus := range cb.appMap {
		if appStatus.isConfigured != appStatus.isSelected {
			appNames = append(appNames, appName)
		}
	}
	return appNames
}

func (cb *appComboBox) onChange(i int) {
	if i == 0 {
		return
	}

	appNames := cb.Items()
	if i >= len(appNames) {
		return
	}
	appName := appNames[i]
	appName = strings.TrimPrefix(appName, selectedPrefix)
	// zrób appName z appNames po i
	// pozbądź się tego ptaszka selectedPrefix z przodu

	appStatus := cb.appMap[appName]
	appStatus.isSelected = !appStatus.isSelected

	cb.appMap[appName] = appStatus
	cb.updateAppList()
}

func (cb *appComboBox) haveUserChanges() bool {
	for _, appStatus := range cb.appMap {
		if appStatus.isSelected {
			// first user change - is user change
			return true
		}
	}
	return false
}

func (cb *appComboBox) populateAppList(appNames []string) {
	for _, appName := range appNames {
		if _, exists := cb.appMap[appName]; !exists {
			cb.appMap[appName] = appComboStatus{}
		}
	}
	cb.updateAppList()
}

func (cb *appComboBox) populateAppConfigStatus(appNames []string) {
	for _, appName := range appNames {
		appStatus := cb.appMap[appName]
		// TODO: to think about
		appStatus.isConfigured = true
		cb.appMap[appName] = appStatus
	}
	cb.updateAppList()
}

func (cb *appComboBox) updateAppList() {
	appNames := make([]string, len(cb.appMap)+1)
	marked := make([]string, 0, len(cb.appMap))

	appKeys := maps.Keys(cb.appMap)
	sort.Strings(appKeys)

	for i, appName := range appKeys {
		// configured but not clicked => check
		// not configured but clicked => check
		// configured and cliucked => unmark
		appStatus := cb.appMap[appName]
		if appStatus.isConfigured != appStatus.isSelected {
			marked = append(marked, appName)
			appName = selectedPrefix + appName
		}

		// entry 0 is status
		appNames[i+1] = appName
	}

	switch l := len(marked); {
	case l == 0:
		appNames[0] = "= Nothing ="
	case l < 4:
		appNames[0] = strings.Join(marked, " & ")
	default:
		appNames[0] = fmt.Sprintf("%d app selected", len(marked))
	}

	cb.SetItems(appNames)
	cb.SetSelectedIndex(0)
}
