package debug

import (
	"encoding/json"
	"fmt"
)

// DeferredJSON is a helper that delays JSON formatting until/unless it is needed.
type DeferredJSON struct {
	O interface{}
}

// String is the method that is called to format the object.
func (d DeferredJSON) String() string {
	b, err := json.Marshal(d.O)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(b)
}

// JSON is a helper that prints the object in JSON format.
// We use lazy-evaluation to avoid calling json.Marshal unless it is actually needed.
func JSON(o interface{}) DeferredJSON {
	return DeferredJSON{o}
}
