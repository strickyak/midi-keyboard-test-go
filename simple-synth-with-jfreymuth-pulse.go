// +build main

/*
  Does not work very well.  How to configure it?
  Maybe I cannot use goroutines; scheduling is too flaky?

  go run simple-synth-with-jfreymuth-pulse.go --rate 16000 --buf 300 -chan 100
*/
package main

import (
	"flag"
	"log"
	"math"
	"os"

	"github.com/jfreymuth/pulse"
)

var Device = flag.String("dev", "/dev/midi1", "midi device name")
var Rate = flag.Int("rate", 48000, "samples per second")
var BufferSize = flag.Int("buf", 100, "samples")
var ChannelSize = flag.Int("chan", 100, "samples")

type Midi struct {
	A, B, C byte
}

func MidiKeyboardTo(out chan Midi) {
	r, err := os.Open(*Device)
	if err != nil {
		log.Fatalf("Cannot open %q: %v", *Device, err)
	}

	bb := make([]byte, 3)
	for {
		n, err := r.Read(bb)
		if err != nil {
			log.Fatalf("Cannot read %q: %v", *Device, err)
		}
		if n != 3 {
			log.Fatalf("Short read %q: %d", *Device, n)
		}
		println("#midi#", bb[0], bb[1], bb[2])
		out <- Midi{bb[0], bb[1], bb[2]}
	}
}

func Synthesize(in chan Midi, out chan float64) {
	var phase, delta, freq, gain float64
	tau := 2 * math.Pi
	rate := float64(*Rate)

	for {
		select {
		case m := <-in:
			switch m.A {
			case 144: // key down
				gain = 1
				freq = 440 * math.Pow(2, float64(int(m.B)-57)/12)
				delta = freq / rate

			case 128: // key up
				gain = 0

			default:
				// nop
			}

		default:
			// nop
		}

		phase += delta
		volt := gain * math.Sin(phase*tau)
		out <- volt
	}
}

func MakeProducer() func(out []float32) {
	midis := make(chan Midi, 10)
	volts := make(chan float64, *ChannelSize)
	go Synthesize(midis, volts)
	go MidiKeyboardTo(midis)

	return func(out []float32) {
		for i, _ := range out {
			out[i] = float32(<-volts) * 0.3
		}
	}
}

func main() {
	flag.Parse()

	c, err := pulse.NewClient()
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	s, err := c.DefaultSink()
	if err != nil {
		log.Println(err)
		return
	}

	fn := MakeProducer()
	stream, err := c.NewPlayback(fn, pulse.PlaybackMono, pulse.PlaybackSampleRate(*Rate), pulse.PlaybackBufferSize(*BufferSize), pulse.PlaybackLowLatency(s))
	if err != nil {
		log.Println(err)
		return
	}

	stream.Start()

	log.Println("Press enter to stop...")
	os.Stdin.Read([]byte{0})

	stream.Close()
}
