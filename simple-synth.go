// +build main

/*
$ go run simple-synth.go | paplay --rate=48000 --channels=1 --format=s16le --latency-msec=20 --raw /dev/stdin
*/
package main

import (
	"flag"
	"log"
	"math"
	"os"
)

var Device = flag.String("dev", "/dev/midi1", "midi device name")
var Rate = flag.Float64("rate", 48000, "samples per second")

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

	for {
		select {
		case m := <-in:
			switch m.A {
			case 144: // key down
				gain = 1
				freq = 440 * math.Pow(2, float64(int(m.B)-57)/12)
				delta = freq / *Rate

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

func main() {
	flag.Parse()
	midis := make(chan Midi, 100)
	volts := make(chan float64, 100)
	pcm := make([]byte, 2)

	go Synthesize(midis, volts)
	go MidiKeyboardTo(midis)

	for {
		volt := <-volts

		x := int16(volt * 10000)
		if x < 0 {
			pcm[0] = byte(x)
			pcm[1] = byte(x >> 8)
		} else {
			pcm[0] = byte(x)
			pcm[1] = byte(x>>8) | 0x80
		}
		n, err := os.Stdout.Write(pcm)
		if err != nil {
			log.Fatalf("Cannot write stdout: %v", err)
		}
		if n != 2 {
			log.Fatalf("short write stdout: %d", n)
		}
	}
}
