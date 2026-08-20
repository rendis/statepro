package experimental

import (
	"testing"
	"unicode/utf8"
)

// FuzzEventBuilder ensures builders tolerate arbitrary names/data without panicking.
func FuzzEventBuilder(f *testing.F) {
	f.Add("go", "k", "v")
	f.Add("", "", "")
	f.Add("ping", "blob", string(make([]byte, 1024)))
	f.Add("🎉", "x", "y")

	f.Fuzz(func(t *testing.T, name, key, val string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("EventBuilder panicked: %v", r)
			}
		}()
		if !utf8.ValidString(name) || !utf8.ValidString(key) {
			return
		}
		evt := NewEventBuilder(name).SetData(map[string]any{key: val}).Build()
		_ = evt.GetEventName()
		_ = evt.GetData()
		_ = evt.DataContainsKey(key)
		_ = evt.GetEvtType()
		_ = evt.GetFlags()
	})
}
