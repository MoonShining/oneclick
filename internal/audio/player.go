package audio

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/youpy/go-wav"
)

type Player struct {
	context *oto.Context
	player  oto.Player
	mu      sync.Mutex
	volume  float64
	current *Audio
	currentReader *byteReader
}

type Audio struct {
	Path     string
	Duration float64
	samples  []byte
}

func NewPlayer(initialVolumePercent float64) (*Player, error) {
	c, ready, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		return nil, err
	}
	<-ready

	// Config stores 0-100, convert to 0.0-1.0
	return &Player{
		context: c,
		volume:  initialVolumePercent / 100.0,
	}, nil
}

func (p *Player) SetVolume(volumePercent float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Config stores 0-100, player uses 0.0-1.0
	p.volume = volumePercent / 100.0
	if p.player != nil {
		p.player.SetVolume(p.volume)
	}
}

func (p *Player) Volume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

func (p *Player) LoadMP3(path string) (*Audio, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return nil, err
	}

	// Read all samples into memory
	samples := make([]byte, 0, d.Length())
	buf := make([]byte, 4096)
	for {
		n, err := d.Read(buf)
		if n > 0 {
			samples = append(samples, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Calculate duration in seconds
	// Sample rate is 44100Hz, 2 channels, 2 bytes per sample
	bytesPerSecond := 44100 * 2 * 2
	duration := float64(len(samples)) / float64(bytesPerSecond)

	return &Audio{
		Path:     path,
		Duration: duration,
		samples:  samples,
	}, nil
}

func (p *Player) LoadWAV(path string) (*Audio, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := wav.NewReader(f)
	format, err := reader.Format()
	if err != nil {
		return nil, err
	}

	// Read all samples into memory as bytes
	var samples []byte
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			samples = append(samples, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	// Calculate duration in seconds
	// SampleRate * NumChannels * BitsPerSample/8 = bytes per second
	bytesPerSecond := uint32(format.SampleRate * uint32(format.NumChannels) * uint32(format.BitsPerSample) / 8)
	duration := float64(len(samples)) / float64(bytesPerSecond)

	return &Audio{
		Path:     path,
		Duration: duration,
		samples:  samples,
	}, nil
}

func (p *Player) Play(a *Audio) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player != nil {
		p.player.Pause()
		p.player = nil
	}

	r := newByteReader(a.samples)
	p.currentReader = r
	player := p.context.NewPlayer(r)
	player.SetVolume(p.volume)

	p.current = a
	p.player = player
	player.Play()

	return nil
}

func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player != nil {
		p.player.Pause()
		p.player = nil
		p.current = nil
		p.currentReader = nil
	}
}

func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player != nil {
		p.player.Pause()
	}
}

func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player != nil {
		p.player.Play()
	}
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.player != nil && p.player.IsPlaying()
}

func (p *Player) Current() *Audio {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *Player) CurrentPosition() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player == nil || p.current == nil || p.currentReader == nil {
		return 0
	}

	return float64(p.currentReader.Position()) / (44100 * 2 * 2)
}

func (p *Player) Seek(pos float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.player == nil || p.current == nil || p.currentReader == nil {
		return
	}

	bytePos := int(pos * 44100 * 2 * 2)
	if bytePos < 0 {
		bytePos = 0
	}
	if bytePos > len(p.current.samples) {
		bytePos = len(p.current.samples)
	}
	p.currentReader.Seek(bytePos)
}

func (p *Player) Close() error {
	p.Stop()
	// oto/v2 Context doesn't have a Close method
	return nil
}

type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data, pos: 0}
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *byteReader) Position() int {
	return r.pos
}

func (r *byteReader) Seek(pos int) {
	r.pos = pos
}
