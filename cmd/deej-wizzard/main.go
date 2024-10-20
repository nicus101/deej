package main

import "github.com/omriharel/deej/pkg/deej/ui"

type fakeProgramLister []string

func (fak fakeProgramLister) ProgramList() ([]string, error) {
	return fak, nil
}

func main() {
	fakeProgramLister := fakeProgramLister{
		"dupa.exe",
		"hujnik.exe",
		"cycki.exe",
		"= UwU =",
	}
	ui.ShowUI(nil, nil, fakeProgramLister, nil, nil)
}
