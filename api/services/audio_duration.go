package services

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/tcolgate/mp3"
)

func detectMP3DurationSeconds(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close MP3 file %s: %v", filePath, err)
		}
	}()

	dec := mp3.NewDecoder(f)
	var (
		frame    mp3.Frame
		skipped  int
		totalDur time.Duration
	)

	for {
		err := dec.Decode(&frame, &skipped)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		totalDur += frame.Duration()
	}

	if totalDur <= 0 {
		return 0, errors.New("unable to detect mp3 duration")
	}

	seconds := int((totalDur + time.Second - 1) / time.Second)
	duration := seconds
	if duration == 0 {
		duration = 1
	}
	return duration, nil
}
