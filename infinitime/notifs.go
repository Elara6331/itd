package infinitime

var (
	regularNotifHeader = []byte{0x00, 0x01, 0x00}
	callNotifHeader    = []byte{0x03, 0x01, 0x00}
)

// Notify sends a notification to the PineTime using the Alert Notification Service
func (d *Device) Notify(title, body string) error {
	c, err := d.getChar(newAlertChar)
	if err != nil {
		return err
	}

	content := title + "\x00" + body
	_, err = c.WriteWithoutResponse(append(regularNotifHeader, content...))
	return err
}

type CallStatus uint8

const (
	CallStatusDeclined CallStatus = iota
	CallStatusAccepted
	CallStatusMuted
)

// NotifyCall sends a call to the PineTime using the Alert Notification Service,
// then executes fn once the user presses a button on the watch.
func (d *Device) NotifyCall(from string, fn func(CallStatus)) error {
	c, err := d.getChar(newAlertChar)
	if err != nil {
		return err
	}

	_, err = c.WriteWithoutResponse(append(callNotifHeader, from...))
	if err != nil {
		return err
	}

	return watchCharOnce(d, notifEventChar, fn)
}
