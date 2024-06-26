package deej

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSerialIO_handleLine(t *testing.T) {
	type testCase struct {
		expectedValues []float32
		expectMutes    []bool
		givenLine      string
		isInvering     bool
	}

	testCases := map[string]testCase{
		"single-value": {
			expectedValues: []float32{0.12},
			givenLine:      "123\r\n",
			isInvering:     false,
		},
		"double-value": {
			expectedValues: []float32{0.12, 0.44},
			givenLine:      "123|456\r\n",
			isInvering:     false,
		},
		"invalid-first-value": {
			expectedValues: []float32{},
			givenLine:      "9999|123|456\r\n",
			isInvering:     false,
		},
		"invalid-other-value": {
			expectedValues: []float32{0.12, 0.44, 9.77},
			givenLine:      "123|456|9999\r\n",
			isInvering:     false,
		},
		"single-value-inverted": {
			expectedValues: []float32{0.88},
			givenLine:      "123\r\n",
			isInvering:     true,
		},
		"gibrish-values": {
			expectedValues: []float32{},
			givenLine:      "UwU",
			isInvering:     false,
		},
		"two-buttons-only": {
			expectMutes:    []bool{true, false},
			expectedValues: []float32{},
			givenLine:      "but|1|0\r\n",
			isInvering:     false,
		},
		"button-line-cut-before-end": {
			expectMutes:    []bool{},
			expectedValues: []float32{},
			givenLine:      "but|0|0|0|0|0|0|0|0|0|0|0|0|0|0|0",
			isInvering:     false,
		},
		"malformed-entries": {
			expectMutes:    []bool{false, false, false, false},
			expectedValues: []float32{},
			givenLine:      "but|wlazl|kotek|na|plotek\r\n",
			isInvering:     false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {

			fak := &fakeMuteConsumer{}
			sio := SerialIO{
				logger: zap.S(),
				deej: &Deej{
					config: &CanonicalConfig{
						InvertSliders: testCase.isInvering,
					},
				},
				sliderMoveConsumers: []chan SliderMoveEvent{
					make(chan SliderMoveEvent, len(testCase.expectedValues)),
				},
				muteConsumer: fak,
			}
			sio.handleLine(zap.S(), testCase.givenLine)

			for i, expectedValue := range testCase.expectedValues {
				sliderEvent := <-sio.sliderMoveConsumers[0]

				assert.Equal(t, i, sliderEvent.SliderID)
				assert.Equal(t, expectedValue, sliderEvent.PercentValue)
			}

			if testCase.expectMutes == nil {
				assert.Nil(t, fak.data, "we expect ")
			}
		})
	}
}

type fakeMuteConsumer struct {
	sync.Mutex
	data []bool
}

func (fak *fakeMuteConsumer) Mute(data []bool) {
	fak.Lock()
	defer fak.Unlock()

	fak.data = data
}
