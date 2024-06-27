// Package deej provides a machine-side client that pairs with an Arduino
// chip to form a tactile, physical volume control system/
package deej

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/omriharel/deej/pkg/deej/ui"
	"github.com/omriharel/deej/pkg/deej/util"
	"github.com/omriharel/deej/pkg/device"
)

const (

	// when this is set to anything, deej won't use a tray icon
	envNoTray = "DEEJ_NO_TRAY_ICON"
)

// Deej is the main entity managing access to all sub-components
type Deej struct {
	logger   *zap.SugaredLogger
	notifier Notifier
	config   *CanonicalConfig
	serial   *SerialIO
	sessions *SessionMap

	stopChannel chan bool
	version     string
	verbose     bool

	connection *device.Connection
}

// NewDeej creates a Deej instance
func NewDeej(logger *zap.SugaredLogger, verbose bool) (*Deej, error) {
	logger = logger.Named("deej")

	notifier, err := NewToastNotifier(logger)
	if err != nil {
		logger.Errorw("Failed to create ToastNotifier", "error", err)
		return nil, fmt.Errorf("create new ToastNotifier: %w", err)
	}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		logger.Errorw("Failed to create Config", "error", err)
		return nil, fmt.Errorf("create new Config: %w", err)
	}

	d := &Deej{
		logger:      logger,
		notifier:    notifier,
		config:      config,
		stopChannel: make(chan bool),
		verbose:     verbose,
		connection:  &device.Connection{},
	}

	serial, err := NewSerialIO(d, logger)
	if err != nil {
		logger.Errorw("Failed to create SerialIO", "error", err)
		return nil, fmt.Errorf("create new SerialIO: %w", err)
	}

	d.serial = serial

	sessionFinder, err := newSessionFinder(logger)
	if err != nil {
		logger.Errorw("Failed to create SessionFinder", "error", err)
		return nil, fmt.Errorf("create new SessionFinder: %w", err)
	}

	sessions, err := newSessionMap(d, logger, sessionFinder)
	if err != nil {
		logger.Errorw("Failed to create sessionMap", "error", err)
		return nil, fmt.Errorf("create new sessionMap: %w", err)
	}

	d.sessions = sessions

	logger.Debug("Created deej instance")

	return d, nil
}

// Initialize sets up components and starts to run in the background
func (d *Deej) Initialize() error {
	d.logger.Debug("Initializing")

	// load the config for the first time
	if err := d.config.Load(); err != nil {
		d.logger.Errorw("Failed to load config during initialization", "error", err)
		return fmt.Errorf("load config during init: %w", err)
	}

	// initialize the session map
	if err := d.sessions.initialize(); err != nil {
		d.logger.Errorw("Failed to initialize session map", "error", err)
		return fmt.Errorf("init session map: %w", err)
	}

	// decide whether to run with/without tray
	if _, noTraySet := os.LookupEnv(envNoTray); noTraySet {

		d.logger.Debugw("Running without tray icon", "reason", "envvar set")

		// run in main thread while waiting on ctrl+C
		d.setupInterruptHandler()
		d.run()

	} else {
		d.setupInterruptHandler()
		d.initializeTray(d.run)
	}

	return nil
}

// TODO: Nicek, opisz to po swojemu
// Z punktu widzenia UI nie chcemy samych sesji, a listę nazw aplikacji/kanałów
// raz że slica interfejsów nie przecastujemy, dwa że vipera i tak obchodzi tylko firefox.exe a nie sesja
func (sf *Deej) ProgramList() ([]string, error) {
	sf.sessions.refreshSessions(true)
	programMap := map[string]struct{}{}

	appendSessionKeys := func(sessions []Session) {
		for _, session := range sessions {
			programName := session.Key()
			programMap[programName] = struct{}{}
		}
	}

	appendSessionKeys(sf.sessions.unmappedSessions)

	for _, sessions := range sf.sessions.m {
		appendSessionKeys(sessions)
	}

	programList := make([]string, 0, len(programMap))
	for programName := range programMap {
		programList = append(programList, programName)
	}

	sort.Strings(programList)
	return programList, nil
}

// SetVersion causes deej to add a version string to its tray menu if called before Initialize
func (d *Deej) SetVersion(version string) {
	d.version = version
}

// Verbose returns a boolean indicating whether deej is running in verbose mode
func (d *Deej) Verbose() bool {
	return d.verbose
}

func (d *Deej) setupInterruptHandler() {
	interruptChannel := util.SetupCloseHandler()

	go func() {
		signal := <-interruptChannel
		d.logger.Debugw("Interrupted", "signal", signal)
		d.signalStop()
	}()
}

func (d *Deej) DevicePortSet(deviceName string) {
	d.connection.DevicePortSet(deviceName)
	fmt.Println("\033[31;1;4mUwU\033[0m")
	d.config.userConfig.Set(configKeyCOMPort, deviceName)
}

func (d *Deej) run() {
	d.logger.Info("Run loop starting")

	// watch the config file for changes
	go d.config.WatchConfigFileChanges()

	// connect to the arduino for the first time
	go func() {
		comPort := d.config.ConnectionInfo.COMPort
		//var lock sync.Mutex
		infoWindowShown := false
		for {
			err := d.connection.ConnectAndDispatch(context.TODO(), comPort, d.sessions) // TODO: make
			if err != nil {
				log.Print("connection failed:", err)
			}
			if !infoWindowShown {
				ui.ConfigInfo()
				infoWindowShown = true
			}
			time.Sleep(3 * time.Second)
			// tutaj nic nie słucha na port change
			// go func() {
			// 	if !lock.TryLock() {
			// 		return
			// 	}
			// 	defer lock.Unlock()

			// 	ui.ShowUI(nil, d, d, d.config, d.config)
			// }()
		}
		// if err := d.serial.Start(); err != nil {
		// 	d.logger.Warnw("Failed to start first-time serial connection", "error", err)

		// 	// If the port is busy, that's because something else is connected - notify and quit
		// 	if errors.Is(err, os.ErrPermission) {
		// 		d.logger.Warnw("Serial port seems busy, notifying user and closing",
		// 			"comPort", d.config.ConnectionInfo.COMPort)

		// 		d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.ConnectionInfo.COMPort),
		// 			"This serial port is busy, make sure to close any serial monitor or other deej instance.")

		// 		d.signalStop()

		// 		// also notify if the COM port they gave isn't found, maybe their config is wrong
		// 	} else if errors.Is(err, os.ErrNotExist) {
		// 		d.logger.Warnw("Provided COM port seems wrong, notifying user and closing",
		// 			"comPort", d.config.ConnectionInfo.COMPort)

		// 		d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.ConnectionInfo.COMPort),
		// 			"This serial port doesn't exist, check your configuration and make sure it's set correctly.")

		// 		d.signalStop()
		// 	}
		// }
	}()

	// wait until stopped (gracefully)
	<-d.stopChannel
	d.logger.Debug("Stop channel signaled, terminating")

	if err := d.stop(); err != nil {
		d.logger.Warnw("Failed to stop deej", "error", err)
		os.Exit(1)
	} else {
		// exit with 0
		os.Exit(0)
	}
}

func (d *Deej) signalStop() {
	d.logger.Debug("Signalling stop channel")
	d.stopChannel <- true
}

func (d *Deej) stop() error {
	d.logger.Info("Stopping")

	d.config.StopWatchingConfigFile()
	d.serial.Stop()

	// release the session map
	if err := d.sessions.release(); err != nil {
		d.logger.Errorw("Failed to release session map", "error", err)
		return fmt.Errorf("release session map: %w", err)
	}

	d.stopTray()

	// attempt to sync on exit - this won't necessarily work but can't harm
	d.logger.Sync()

	return nil
}
