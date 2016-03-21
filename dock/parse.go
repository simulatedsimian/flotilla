package dock

import (
	"bytes"
	"fmt"
)

func makeMessageSplitter() func(input []byte) []string {
	buffer := []byte{}

	return func(input []byte) []string {
		if len(buffer) > 0 && buffer[0] == '\n' {
			buffer = buffer[1:]
		}

		buffer = append(buffer, input...)
		msgs := []string{}

		for {
			i := bytes.IndexByte(buffer, '\r')
			if i == -1 {
				break
			}
			msgs = append(msgs, string(buffer[:i]))
			buffer = buffer[i+1:]
			// skip past any \n that after the \r
			if len(buffer) > 0 && buffer[0] == '\n' {
				buffer = buffer[1:]
			}
		}

		return msgs
	}
}

// u 1/module int,int,int,int....
func msgToEvent(msg string) Event {
	var evtype rune
	var port int
	var module string
	var values [8]int

	n, _ := fmt.Sscanf(msg, "%c %d/%s %d,%d,%d,%d,%d,%d,%d,%d",
		&evtype, &port, &module,
		&values[0], &values[1], &values[2], &values[3],
		&values[4], &values[5], &values[6], &values[7])

	if evtype == '#' {
		return Event{EventType: Message, Message: msg}
	}

	event := Event{EventType: Invalid}

	if n >= 3 {
		switch evtype {
		case 'c':
			event.EventType = Connected
		case 'd':
			event.EventType = Disconnected
		case 'u':
			event.EventType = Update
		}

		event.ModuleType = FromString(module)
		event.Port = port
		if n > 3 {
			event.Params = values[:n-3]
		}
	}
	return event
}

func join(a []int, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return fmt.Sprint(a[0])
	}

	var b bytes.Buffer
	for i, v := range a {
		fmt.Fprint(&b, v)
		if i < (len(a) - 1) {
			fmt.Fprint(&b, sep)
		}
	}

	return b.String()
}
