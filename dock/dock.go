package dock

import (
	"fmt"
	"io"
	"log"
	"sync"
)

type Dock struct {
	port   io.ReadWriter
	Events chan Event
	sync.RWMutex
	moduleTypes map[int]ModuleType
}

func ConnectDock(port io.ReadWriter) *Dock {
	dock := Dock{
		port:        port,
		Events:      make(chan Event, 128),
		moduleTypes: make(map[int]ModuleType),
	}

	go dock.reader()

	return &dock
}

func (d *Dock) handleEvent(ev Event) {

	log.Println(ev)

	if ev.EventType == EventDisconnected {
		d.RWMutex.Lock()
		d.moduleTypes[ev.Channel] = Unknown
		d.RWMutex.Unlock()
	}

	if ev.EventType == EventConnected {
		d.RWMutex.Lock()
		d.moduleTypes[ev.Channel] = ev.ModuleType
		d.RWMutex.Unlock()
	}

	d.Events <- ev
}

func (d *Dock) reader() {
	splitter := makeMessageSplitter([]byte{0x0d, 0x0a}) // cr lf
	buffer := make([]byte, 128)

	for {
		n, err := d.port.Read(buffer)
		if n > 0 {
			msgs := splitter(buffer[:n])
			for _, msg := range msgs {
				d.handleEvent(msgToEvent(msg))
			}
		}

		if err != nil {
			d.handleEvent(Event{EventType: EventError, Error: err})
			return
		}
	}
}

func (d *Dock) SendDockCommand(command rune, params ...int) error {

	if len(params) == 0 {
		_, err := fmt.Fprintf(d.port, "%c\r", command)
		return err
	}
	_, err := fmt.Fprintf(d.port, "%c %s\r", command, join(params, ","))
	return err
}

func (d *Dock) SetModuleData(port int, mtype ModuleType, params ...int) error {
	err := validateParams(mtype, params)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(d.port, "s %d %s\r", port, join(params, ","))
	return err
}

func (d *Dock) GetModuleType(port int) ModuleType {
	d.RWMutex.RLock()
	defer d.RWMutex.RUnlock()
	if mt, ok := d.moduleTypes[port]; ok {
		return mt
	}
	return Unknown
}
