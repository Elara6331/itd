package mpris

import (
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestParsePropertiesChanged checks the parsePropertiesChanged function to
// make sure it correctly parses a DBus PropertiesChanged signal.
func TestParsePropertiesChanged(t *testing.T) {
	// Create a DBus message
	msg := &dbus.Message{
		Body: []interface{}{
			"com.example.Interface",
			map[string]dbus.Variant{
				"Property1": dbus.MakeVariant(true),
				"Property2": dbus.MakeVariant("Hello, world!"),
			},
			[]string{},
		},
	}

	// Parse the message
	iface, changed, ok := parsePropertiesChanged(msg)
	if !ok {
		t.Error("Expected parsePropertiesChanged to return true, but got false")
	}

	// Check the parsed values
	expectedIface := "com.example.Interface"
	if iface != expectedIface {
		t.Errorf("Expected iface to be %q, but got %q", expectedIface, iface)
	}

	expectedChanged := map[string]dbus.Variant{
		"Property1": dbus.MakeVariant(true),
		"Property2": dbus.MakeVariant("Hello, world!"),
	}
	if !reflect.DeepEqual(changed, expectedChanged) {
		t.Errorf("Expected changed to be %v, but got %v", expectedChanged, changed)
	}

	// Test a message with an invalid number of arguments
	msg = &dbus.Message{
		Body: []interface{}{
			"com.example.Interface",
		},
	}
	_, _, ok = parsePropertiesChanged(msg)
	if ok {
		t.Error("Expected parsePropertiesChanged to return false, but got true")
	}

	// Test a message with an invalid first argument
	msg = &dbus.Message{
		Body: []interface{}{
			123,
			map[string]dbus.Variant{
				"Property1": dbus.MakeVariant(true),
				"Property2": dbus.MakeVariant("Hello, world!"),
			},
			[]string{},
		},
	}
	_, _, ok = parsePropertiesChanged(msg)
	if ok {
		t.Error("Expected parsePropertiesChanged to return false, but got true")
	}

	// Test a message with an invalid second argument
	msg = &dbus.Message{
		Body: []interface{}{
			"com.example.Interface",
			123,
			[]string{},
		},
	}
	_, _, ok = parsePropertiesChanged(msg)
	if ok {
		t.Error("Expected parsePropertiesChanged to return false, but got true")
	}
}
