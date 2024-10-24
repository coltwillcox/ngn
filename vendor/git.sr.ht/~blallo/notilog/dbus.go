package notilog

import (
	"fmt"

	"github.com/godbus/dbus/v5"

	"git.sr.ht/~blallo/conductor"
)

type Action int

const (
	StopAction Action = iota
	RestartAction
)

type NotiListener struct {
	conn *dbus.Conn
	out  chan *dbus.Message
}

func NewNotiListener(out chan *dbus.Message) (*NotiListener, error) {
	conn, err := newHookedConn()
	if err != nil {
		return nil, err
	}

	return &NotiListener{
		conn: conn,
		out:  out,
	}, nil
}

func (n *NotiListener) Close() error {
	// NOTE: this also closes the n.out channel
	return n.conn.Close()
}

func (n *NotiListener) Run(c conductor.Conductor[Action]) error {
	n.conn.Eavesdrop(n.out)

	for {
		select {
		case cmd := <-c.Cmd():
			switch cmd {
			case StopAction:
				return n.Close()
			case RestartAction:
				conn, err := newHookedConn()
				if err != nil {
					return err
				}
				// XXX: Calling Close on the old connection also closes
				// the out channel. This makes the main loop exit.
				// Just hope that the old connection is garbage-collected,
				// or we leak resources.
				n.conn = conn
				// XXX: This call is recursive. I don't know if the go compiler
				// optimizes for tail calls, but after a high enough amount of
				// restarts we might have a stack overflow.
				return n.Run(c)
			}
		case <-c.Done():
			return n.Close()
		}
	}
}

func newHookedConn() (*dbus.Conn, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	call := conn.BusObject().Call(
		"org.freedesktop.DBus.Monitoring.BecomeMonitor", // The interface
		1, // The call does not return any value
		[]string{"eavesdrop=true,interface='org.freedesktop.Notifications',member='Notify'"}, // What we want to listen to
		uint32(0), // Unused but must be uint32(0) ¯\_(ツ)_/¯
	)
	if call.Err != nil {
		// As the function does not return, these are most probably errors
		// while communicating with dbus.
		return nil, fmt.Errorf("failed to become monitor: %w", call.Err)
	}

	return conn, nil
}
