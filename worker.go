package sanic

import (
	"log"
	"sync"
	"time"
)

type Worker struct {
	ID             int64 // 0 - 2 ^ IDBits
	IDBits         uint64
	IDShift        uint64
	Sequence       int64 // 0 - 2 ^ SequenceBits
	SequenceBits   uint64
	LastTimeStamp  int64
	TimeStampBits  uint64
	TimeStampShift uint64
	Frequency      time.Duration
	TotalBits      uint64
	CustomEpoch    int64
	mutex          sync.Mutex
}

func NewWorker(
	id, epoch int64, idBits, sequenceBits, timestampBits uint64,
	frequency time.Duration) Worker {

	totalBits := idBits + sequenceBits + timestampBits + 1
	if totalBits%6 != 0 {
		log.Fatal("totalBits + 1 must be evenly divisible by 6")
	}

	w := Worker{
		ID:             id,
		IDBits:         idBits,
		IDShift:        sequenceBits,
		Sequence:       0,
		SequenceBits:   sequenceBits,
		TimeStampBits:  timestampBits,
		TimeStampShift: sequenceBits + idBits,
		Frequency:      frequency,
		TotalBits:      totalBits,
		CustomEpoch:    epoch,
	}
	w.LastTimeStamp = w.Time()
	return w
}

// Pre-generated workers are usable examples set as the 0'th worker of their
// type, with a custom epoch of "2016-01-01 00:00:00 +0000 UTC"

// TenLengthWorker will generate up to 8192000 unique ids/second for 69 years
var TenLengthWorker = NewWorker(0, 1451606400000, 5, 13, 41, time.Millisecond)

// NineLengthWorker will generate up to 819200 unique ids/second for 87 years
var NineLengthWorker = NewWorker(0, 145160640000, 2, 13, 38, time.Millisecond*10)

// EightLengthWorker will generate up to 40960 unique ids/second for 54 years
var EightLengthWorker = NewWorker(0, 14516064000, 1, 12, 34, time.Millisecond*100)

// SevenLengthWorker will generate up to 1024 unique ids/second for 68 years
var SevenLengthWorker = NewWorker(0, 1451606400, 0, 10, 31, time.Second)

func (w *Worker) NextID() int64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.UnsafeNextID()
}

// UnsafeNextID is faster than NextID, but must be called within
// only one goroutine, otherwise ID uniqueness is not guaranteed.
func (w *Worker) UnsafeNextID() int64 {
	timestamp := w.Time()

	if w.LastTimeStamp > timestamp {
		w.waitForNextTime()
	}

	if w.LastTimeStamp == timestamp {
		w.Sequence = (w.Sequence + 1) % (1 << w.SequenceBits)
		if w.Sequence == 0 {
			w.waitForNextTime()
			timestamp = w.LastTimeStamp
		}
	} else {
		w.Sequence = 0
	}

	w.LastTimeStamp = timestamp

	return (timestamp-w.CustomEpoch)<<w.TimeStampShift |
		w.ID<<w.IDShift |
		w.Sequence
}

func (w *Worker) IDString(id int64) string {
	str, _ := IntToString(id, w.TotalBits)
	return str
}

func (w *Worker) waitForNextTime() {
	ts := w.Time()
	for ts <= w.LastTimeStamp {
		ts = w.Time()
	}
	w.LastTimeStamp = ts
}

func (w *Worker) Time() int64 {
	return time.Now().UnixNano() / int64(w.Frequency)
}
