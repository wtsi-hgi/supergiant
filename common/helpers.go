package common

import (
	"fmt"
	"time"
)

func WaitFor(desc string, d time.Duration, i time.Duration, fn func() (bool, error)) error {
	started := time.Now()
	for {
		elapsed := time.Since(started)
		if elapsed > d {
			return fmt.Errorf("Timed out waiting for %s", desc)
		}
		if ok, err := fn(); ok {
			return nil
		} else if err != nil {
			return err
		}
		time.Sleep(i)
	}
}

func StringID(id ID) string {
	if id == nil {
		panic("Attempting pointer dereference on nil Resource ID field")
	}
	return *id
}

func IDString(str string) ID {
	return &str
}
