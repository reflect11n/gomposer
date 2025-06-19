package window

import "github.com/reflect11n/gomposer/window/interfaces"

type Window struct {
	interfaces.Window
	Width      int16
	Height     int16
	IsActive   bool
	StayActive bool
	ReadOnly   bool
}
