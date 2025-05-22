package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	d.Duration = dur
	return nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	s := string(text)
	if dur, err := time.ParseDuration(s); err == nil {
		d.Duration = dur
		return nil
	}
	if min, err := strconv.Atoi(s); err == nil {
		d.Duration = time.Duration(min) * time.Minute
		return nil
	}
	return fmt.Errorf("invalid duration format: %s", s)
}
