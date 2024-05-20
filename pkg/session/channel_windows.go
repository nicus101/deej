package session

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/mitchellh/go-ps"
	"github.com/moutend/go-wca"
)

type Channel struct {
	lock sync.Mutex

	sysAudio *wca.ISimpleAudioVolume
	sysGUID  *ole.GUID
	process  ps.Process

	// lastMute   bool
	// lastVolume float32
}

// TODO: make private
func (ch *Channel) Release() {
	ch.sysAudio.Release()
}

func (ch *Channel) SetVolume(volume float32) (err error) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	err = ch.sysAudio.SetMasterVolume(volume, ch.sysGUID)
	if err != nil {
		return fmt.Errorf("adjust session volume: %w", err)
	}

	// TODO: state checking
	// TODO: updating last state

	// // mitigate expired sessions by checking the state whenever we change volumes
	// var state uint32

	// if err := s.control.GetState(&state); err != nil {
	// 	s.logger.Warnw("Failed to get session state while setting volume", "error", err)
	// 	return fmt.Errorf("get session state: %w", err)
	// }

	// if state == wca.AudioSessionStateExpired {
	// 	s.logger.Warnw("Audio session expired, triggering session refresh")
	// 	return errRefreshSessions
	// }

	// s.logger.Debugw("Adjusting session volume", "to", fmt.Sprintf("%.2f", v))

	return nil
}

func (ch *Channel) SetMute(mute bool) (err error) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// TODO: check if change needed

	err = ch.sysAudio.SetMute(mute, ch.sysGUID)
	if err != nil {
		return fmt.Errorf("adjust session mute: %w", err)
	}

	// TODO: state checking
	// TODO: updating last state

	// // mitigate expired sessions by checking the state whenever we change volumes
	// var state uint32

	// if err := s.control.GetState(&state); err != nil {
	// 	s.logger.Warnw("Failed to get session state while setting volume", "error", err)
	// 	return fmt.Errorf("get session state: %w", err)
	// }

	// if state == wca.AudioSessionStateExpired {
	// 	s.logger.Warnw("Audio session expired, triggering session refresh")
	// 	return errRefreshSessions
	// }

	// s.logger.Debugw("Adjusting session mute", "to", m)
	// s.isMuted = m
	return nil
}

func (ch *Channel) Executable() string {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	return ch.process.Executable()
}

// TODO: make private
func MakeChannel(audioSessionControl *wca.IAudioSessionControl, sysGUID *ole.GUID) (*Channel, error) {

	// query its IAudioSessionControl2
	dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
	if err != nil {
		return nil, fmt.Errorf("iAudioSessionControl2: %w", err)
	}
	// receive a useful object instead of our dispatch
	audioSessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))

	var pid uint32
	// get the session's PID
	if err := audioSessionControl2.GetProcessId(&pid); err != nil {
		// if this is the system sounds session, GetProcessId will error with an undocumented
		// AUDCLNT_S_NO_CURRENT_PROCESS (0x889000D) - this is fine, we actually want to treat it a bit differently
		// The first part of this condition will be true if the call to IsSystemSoundsSession fails
		// The second part will be true if the original error mesage from GetProcessId doesn't contain this magical
		// error code (in decimal format).
		isSystemSoundsErr := audioSessionControl2.IsSystemSoundsSession()
		if isSystemSoundsErr != nil && !strings.Contains(err.Error(), "143196173") {
			return nil, fmt.Errorf("obtaining pid: %w", err)
		}
		// update 2020/08/31: this is also the exact case for UWP applications, so we should no longer override the PID.
		// it will successfully update whenever we call GetProcessId for e.g. Video.UI.exe, despite the error being non-nil.
	}

	// get its ISimpleAudioVolume
	dispatch, err = audioSessionControl2.QueryInterface(wca.IID_ISimpleAudioVolume)
	if err != nil {
		return nil, fmt.Errorf("cannot recast to ISimpleAudioVolume: %w", err)
	}
	// make it useful, again
	simpleAudioVolume := (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch))

	process, err := ps.FindProcess(int(pid))
	if err != nil {
		return nil, fmt.Errorf("find process name by pid: %w", err)
	}

	return &Channel{
		process:  process,
		sysAudio: simpleAudioVolume,
		sysGUID:  sysGUID,
	}, nil
}
