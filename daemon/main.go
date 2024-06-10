package main

import (
	"errors"
	"fmt"
	"os"

	"git.sr.ht/~blallo/conductor"
	"git.sr.ht/~blallo/logz/zlog"
	"git.sr.ht/~blallo/notilog"
	"github.com/godbus/dbus/v5"
	"go.bug.st/serial"
)

func main() {
	log := zlog.NewConsoleLogger()
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Err(map[string]any{
			"msg": "Failed to get ports",
			"err": err.Error(),
		})
		os.Exit(1)
	}
	if len(ports) == 0 {
		log.Err(map[string]any{
			"msg": "No serial ports found",
		})
		os.Exit(1)
	}
	for _, port := range ports {
		log.Info(map[string]any{
			"msg": fmt.Sprintf("Found port: %v\n", port),
		})
	}

	mode := &serial.Mode{}

	port, err := serial.Open("/dev/ttyACM0", mode)
	if err != nil {
		log.Err(map[string]any{
			"msg": "Failed to open port",
			"err": err.Error(),
		})
		os.Exit(1)
	}

	msgCh := make(chan *dbus.Message, 100)
	c := conductor.Simple[notilog.Action]()

	l, err := notilog.NewNotiListener(msgCh)
	if err != nil {
		log.Err(map[string]any{
			"msg": "failed to initialize listener",
			"err": err.Error(),
		})
		os.Exit(1)
	}

	go func() {
		if err := l.Run(c); err != nil {
			log.Err(map[string]any{
				"msg": "execution failed",
				"err": err.Error(),
			})
			os.Exit(1)
		}
		log.Info(map[string]any{"msg": "exited successfully"})
	}()

	for {
		select {
		case msg := <-msgCh:
			if msg == nil {
				log.Info(map[string]any{
					"msg": "channel closed: exiting",
				})
				return
			}
			notification, err := notilog.FromMessage(msg)
			if err != nil {
				if errors.Is(err, notilog.ErrNotANotification) {
					log.Debug(map[string]any{
						"msg": "not a notification",
					})
				} else {
					log.Warn(map[string]any{
						"msg": "failed translating to message",
						"err": err.Error(),
					})
				}
				continue
			}

			log.Info(map[string]any{
				"msg":     "message intercepted",
				"headers": notilog.TranslateHeaders(msg.Headers),
				"content": notification,
			})
			// Send to serial
			_, err = port.Write([]byte(notification.Title))
			if err != nil {
				log.Err(map[string]any{
					"msg": "Failed to write to port",
					"err": err.Error(),
				})
				os.Exit(1)
			}
		}
	}
}
