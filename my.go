package awesomeProject1

import (
	"github.com/undefinedlabs/go-mpatch"
	"reflect"
	"sync"
	"testing"
)

// Pointer of the MonkeyPatching instance
var mpatchPointer *mpatch.Patch

// Map to store the instrumented functions pointers
var funcsMap map[uintptr]bool
var funcsMapMutext sync.Mutex

// Callbacks funcs
var onTestBegin func(*testing.T)
var onTestEnd func(*testing.T)

func InstrumentTests() func() {
	// Initialize functions pointers map
	funcsMap = make(map[uintptr]bool)

	// Get `testing.T.Run` method by reflection
	var t *testing.T
	var err error
	tType := reflect.TypeOf(t)
	tRunMethod, _ := tType.MethodByName("Run")

	// Patch the `Run` method a redirect all calls to `testingTRunModified`
	mpatchPointer, err = mpatch.PatchMethodByReflect(tRunMethod, testingTRunModified)
	if err != nil {
		print(err)
	}

	// Return a func to unpatch and revert the monkey patching changes
	return func() {
		_ = mpatchPointer.Unpatch()
	}
}

func testingTRunModified(instance *testing.T, name string, f func(*testing.T)) bool {
	// Create a signal channel to get the result of the test
	signal := make(chan bool, 1)

	// We need to unpatch and revert the monkey patching to be able to call the original Run func
	err := mpatchPointer.Unpatch()
	if err == nil {
		// We call the original Run func in another goroutine
		go func() {
			signal <- instance.Run(name, getModifiedTestFunc(f))
		}()

		// We restore the patching
		_ = mpatchPointer.Patch()

		// We wait for the signal with the result of the test
		retValue := <-signal
		return retValue
	} else {
		return instance.Run(name, getModifiedTestFunc(f))
	}
}

func getModifiedTestFunc(f func(*testing.T)) func(*testing.T) {
	// Pointer of func is a const, so we don't worry about any GC
	fPtr := reflect.ValueOf(f).Pointer()

	// Lock for map access
	funcsMapMutext.Lock()
	defer funcsMapMutext.Unlock()

	if _, containsFunc := funcsMap[fPtr]; containsFunc {
		// f is already an instrumented func, we just return it
		return f
	} else {
		// We replace the original test function with a wrapper
		f2 := func(t2 *testing.T) {
			// Call onBegin callback
			fireOnTestBegin(t2)
			// Call test func
			f(t2)
			// Call onEnd callback
			fireOnTestEnd(t2)
		}
		// We store the pointer of this function to avoid instrumenting multiple times
		funcsMap[reflect.ValueOf(f2).Pointer()] = true
		return f2
	}
}

func fireOnTestBegin(t *testing.T) {
	// We do a local copy to avoid any threading issue and ensure we are comparing and calling the same callback
	lOnTestBegin := onTestBegin
	if lOnTestBegin != nil {
		lOnTestBegin(t)
	}
}

func fireOnTestEnd(t *testing.T) {
	// We do a local copy to avoid any threading issue and ensure we are comparing and calling the same callback
	lOnTestEnd := onTestEnd
	if lOnTestEnd != nil {
		lOnTestEnd(t)
	}
}

func SetOnTestBegin(fn func(*testing.T)) {
	// Sets the onTestBegin callback
	onTestBegin = fn
}

func SetOnTestEnd(fn func(*testing.T)) {
	// Sets the onTestEnd callback
	onTestEnd = fn
}
