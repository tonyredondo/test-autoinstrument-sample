package awesomeProject1

import (
	"fmt"
	"os"
	"testing"
	_ "unsafe"
)

func TestMain(m *testing.M) {
	defer InstrumentTests()()
	SetOnTestBegin(func(t *testing.T) {
		fmt.Printf("$ test start from instrumentation: %v\n", t.Name())
	})
	SetOnTestEnd(func(t *testing.T) {
		fmt.Printf("$ test end from instrumentation: %v\n", t.Name())
	})
	os.Exit(m.Run())
}

func TestA(t *testing.T) {
	t.Run("my name", func(t2 *testing.T) {
	})
	t.Run("my name2", func(t2 *testing.T) {
	})
	t.Run("my name3", func(t2 *testing.T) {
	})
}

func TestB(t *testing.T) {
	t.Run("my name", func(t2 *testing.T) {
		t2.Run("inner child", func(t3 *testing.T) {
		})
	})
	t.Run("my name2", func(t2 *testing.T) {
	})
	t.Run("my name3", func(t2 *testing.T) {
	})
}
