package audio

import "time"

type Filter interface {
	Apply(audioStream Stream, samples []int16, playTime time.Time, sequence uint32, startTs uint32) (int, error)
}