package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"git.sr.ht/~blallo/conductor"
	logz "git.sr.ht/~blallo/logz/interface"
	"git.sr.ht/~blallo/logz/zlog"
	"git.sr.ht/~blallo/notilog"
	"github.com/godbus/dbus/v5"
	"go.bug.st/serial"
)

func main() {
	log := logFn()

	ports, err := serial.GetPortsList()
	if err != nil {
		log(logz.LogErr, "failed to get ports", err)
	}
	if len(ports) == 0 {
		log(logz.LogErr, "no serial ports found")
	}
	for _, port := range ports {
		log(logz.LogInfo, fmt.Sprintf("found port: %v\n", port))
	}

	port, err := serial.Open("/dev/ttyACM0", &serial.Mode{})
	if err != nil {
		log(logz.LogErr, "failed to open port", err)
	}

	msgCh := make(chan *dbus.Message, 100)
	c := conductor.Simple[notilog.Action]()

	l, err := notilog.NewNotiListener(msgCh)
	if err != nil {
		log(logz.LogErr, "failed to initialize listener", err)
	}

	go func() {
		if err := l.Run(c); err != nil {
			log(logz.LogErr, "execution failed", err)
		}
		log(logz.LogInfo, "exited successfully")
	}()

	for dbusMessage := range msgCh {
		if dbusMessage == nil {
			log(logz.LogInfo, "channel closed: exiting")
			return
		}
		notification, err := notilog.FromMessage(dbusMessage)
		if err != nil {
			if errors.Is(err, notilog.ErrNotANotification) {
				log(logz.LogDebug, "not a notification")
			} else {
				log(logz.LogWarn, "failed translating to message", err)
			}
			continue
		}
		log(logz.LogInfo, fmt.Sprintf("message intercepted: %v\n", notification))

		serialMessage, err := json.Marshal(notification)
		if err != nil {
			log(logz.LogErr, "failed to marshal notification", err)
		}

		// Send to serial
		if _, err = port.Write(serialMessage); err != nil {
			log(logz.LogErr, "failed to write to port", err)
		}
	}
}

func logFn() func(logz.LogLevel, string, ...error) {
	log := zlog.NewConsoleLogger()
	return func(level logz.LogLevel, msg string, errs ...error) {
		data := map[string]any{"msg": msg}
		if len(errs) > 0 && errs[0] != nil {
			data["err"] = errs[0].Error()
		}
		log.Log(level, data)
		if level == logz.LogErr {
			os.Exit(1)
		}
	}
}
