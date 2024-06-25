package device

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnection_DevicePortSet(t *testing.T) {
	// XXX: manual test only
	t.SkipNow()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var connection Connection

	go func() {
		time.Sleep(5 * time.Second)
		connection.DevicePortSet("COM16")
	}()

	for {
		err := connection.ConnectAndDispatch(ctx, "COM3", nil)
		require.ErrorIs(t, err, ErrConnectionTimeout)
	}
}
