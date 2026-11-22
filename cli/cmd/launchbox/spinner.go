package main

import (
	"fmt"
	"sync"
	"time"
)

var spinnerStyles = map[string][]string{
	"arrow":  {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	"dots":   {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"line":   {"|", "/", "-", "\\"},
	"earth":  {"🌍", "🌎", "🌏"},
	"moon":   {"🌕", "🌖", "🌗", "🌘", "🌑", "🌒", "🌓", "🌔"},
	"hearts": {"💛", "💙", "💜", "💚", "❤️"},
}

type Spinner struct {
	mtx    sync.Mutex
	msg    string
	active bool
	stop   chan struct{}
	done   chan struct{}
	style  string
}

func NewSpinner(style string) *Spinner {
	if _, ok := spinnerStyles[style]; !ok {
		fmt.Printf("⚠️ Unknown spinner style: %q. Falling back to \"dots\".\n", style)
		style = "dots"
	}

	return &Spinner{
		style: style,
	}
}

func (s *Spinner) Start(msg string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.active {
		// Replace message only, clean the line
		fmt.Print("\r\033[K")
		s.msg = msg
		return
	}

	s.msg = msg
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	s.active = true

	style := s.style

	go func() {
		defer close(s.done)

		frames := spinnerStyles[style]
		i := 0

		for {
			select {
			case <-s.stop:
				return
			default:
				s.mtx.Lock()
				fmt.Print("\r\033[K")
				fmt.Printf("%s %s", frames[i], s.msg)
				s.mtx.Unlock()

				time.Sleep(100 * time.Millisecond)
				i = (i + 1) % len(frames)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mtx.Lock()
	if !s.active {
		s.mtx.Unlock()
		return
	}

	close(s.stop)
	s.mtx.Unlock() // unlock before waiting

	<-s.done // wait for spinner to stop
	time.Sleep(100 * time.Millisecond)

	s.mtx.Lock()
	s.active = false
	fmt.Print("\r\033[K") // clear line
	s.mtx.Unlock()
}
