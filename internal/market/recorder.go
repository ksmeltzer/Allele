package market

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Recorder manages appending raw Polymarket WebSocket JSON to disk.
type Recorder struct {
	file *os.File
}

// NewRecorder initializes a new data recording file.
func NewRecorder(dir string) (*Recorder, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create a new log file for the day
	filename := filepath.Join(dir, "polymarket_ticks_"+time.Now().Format("2006-01-02")+".jsonl")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	log.Printf("Initialized historical tick recorder: %s", filename)
	return &Recorder{file: file}, nil
}

// WriteTick appends a single raw JSON message to the log file.
func (r *Recorder) WriteTick(msg []byte) error {
	// Ensure the message ends with a newline for JSONL format
	if len(msg) > 0 && msg[len(msg)-1] != '\n' {
		msg = append(msg, '\n')
	}

	_, err := r.file.Write(msg)
	return err
}

// Close safely closes the recording file.
func (r *Recorder) Close() error {
	return r.file.Close()
}
