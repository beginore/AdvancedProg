// test_main.go
package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Chdir("../..")
	code := m.Run()
	os.Exit(code)
}
