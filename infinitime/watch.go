package infinitime

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"

	"tinygo.org/x/bluetooth"
)

type notifier interface {
	notify([]byte)
}

type watcher[T any] struct {
	mu         sync.Mutex
	nextFuncID int
	callbacks  map[int]func(T, error)
	char       *bluetooth.DeviceCharacteristic
}

func (w *watcher[T]) addCallback(fn func(T, error)) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	funcID := w.nextFuncID
	w.callbacks[funcID] = fn
	w.nextFuncID++
	return funcID
}

func (w *watcher[T]) notify(b []byte) {
	var val T
	err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &val)
	w.mu.Lock()
	for _, fn := range w.callbacks {
		go fn(val, err)
	}
	w.mu.Unlock()
}

func (w *watcher[T]) cancelFn(d *Device, ch btChar, id int) func() {
	return func() {
		w.mu.Lock()
		delete(w.callbacks, id)
		w.mu.Unlock()

		if len(w.callbacks) == 0 {
			d.notifierMtx.Lock()
			delete(d.notifierMap, ch)
			d.notifierMtx.Unlock()
			w.char.EnableNotifications(nil)
		}
	}
}

func watchChar[T any](ctx context.Context, d *Device, ch btChar, fn func(T, error)) error {
	d.notifierMtx.Lock()
	defer d.notifierMtx.Unlock()

	if n, ok := d.notifierMap[ch]; ok {
		w := n.(*watcher[T])
		funcID := w.addCallback(fn)
		context.AfterFunc(ctx, w.cancelFn(d, ch, funcID))
		go func() {
			<-ctx.Done()
			w.cancelFn(d, ch, funcID)()
		}()
		return nil
	} else {
		c, err := d.getChar(ch)
		if err != nil {
			return err
		}

		w := &watcher[T]{callbacks: map[int]func(T, error){}}
		err = c.EnableNotifications(w.notify)
		if err != nil {
			return err
		}
		w.char = c
		funcID := w.addCallback(fn)
		d.notifierMap[ch] = w

		context.AfterFunc(ctx, w.cancelFn(d, ch, funcID))
		return nil
	}
}

func watchCharOnce[T any](d *Device, ch btChar, fn func(T)) error {
	ctx, cancel := context.WithCancel(context.Background())

	var watchErr error
	err := watchChar(ctx, d, ch, func(val T, err error) {
		defer cancel()
		if err != nil {
			watchErr = err
			return
		}
		fn(val)
	})
	if err != nil {
		return err
	}

	<-ctx.Done()
	return watchErr
}
