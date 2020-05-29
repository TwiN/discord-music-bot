package dca

import (
	"time"
)

type OpusReader interface {
	OpusFrame() (frame []byte, err error)
	FrameDuration() time.Duration
}
