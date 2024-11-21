// TODO Exit menu option
// TODO Reconnect function
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"fyne.io/systray"
	"git.sr.ht/~blallo/conductor"
	logz "git.sr.ht/~blallo/logz/interface"
	"git.sr.ht/~blallo/logz/zlog"
	"git.sr.ht/~blallo/notilog"
	"github.com/godbus/dbus/v5"
	"go.bug.st/serial"
)

type Notification struct {
	Program   string `json:"program"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Sender    string `json:"sender"`
	Serial    string `json:"serial"`
	CreatedAt string `json:"created_at"`
}

const (
	badgePort = "/dev/ttyACM0"
)

var (
	//go:embed icons/icon-green.png
	iconGreen []byte
	//go:embed icons/icon-red.png
	iconRed []byte
	//go:embed icons/icon-yellow.png
	iconYellow []byte
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconYellow)
	systray.SetTitle("Awesome App")
	systray.SetTooltip("Pretty awesome")
	addExitItem()

	go func() {
		log := logFn()

		msgCh := make(chan *dbus.Message, 100)
		c := conductor.Simple[notilog.Action]()

		l, err := notilog.NewNotiListener(msgCh)
		if err != nil {
			log(logz.LogErr, "failed to initialize listener", err)
			systray.SetIcon(iconRed)
		}

		go func() {
			if err := l.Run(c); err != nil {
				log(logz.LogErr, "execution failed", err)
			}
			log(logz.LogInfo, "exited successfully")
		}()

		connectionChannel := make(chan bool, 1)
		connectionChannel <- true
		for {
			select {
			case <-connectionChannel:
				log(logz.LogInfo, "channel")
				port, err := serial.Open(badgePort, &serial.Mode{})
				if err != nil {
					log(logz.LogErr, "failed to open port", err)
					systray.SetIcon(iconRed)
					time.Sleep(1 * time.Second)
					log(logz.LogInfo, "reconnecting... 1")
					connectionChannel <- true
					break
				}

				// Port existing/connected listener.
				go func() {
					for {
						log(logz.LogInfo, "func start")
						time.Sleep(1 * time.Second)
						portsNames, err := serial.GetPortsList()
						if err != nil {
							log(logz.LogErr, "failed to get ports", err)
							systray.SetIcon(iconRed)
							port.Close()
							time.Sleep(1 * time.Second)
							log(logz.LogInfo, "reconnecting... 2")
							break
						}
						if len(portsNames) == 0 {
							log(logz.LogErr, "no serial ports found")
							systray.SetIcon(iconRed)
							port.Close()
							time.Sleep(1 * time.Second)
							log(logz.LogInfo, "reconnecting... 3")
							break
						}
						existing := false
						for _, portName := range portsNames {
							if portName == badgePort {
								// log(logz.LogInfo, fmt.Sprintf("found port: %v\n", port))
								existing = true
								break
							}
						}
						if !existing {
							log(logz.LogInfo, "not existing. TODO reconnect")
							systray.SetIcon(iconRed)
							port.Close()
							time.Sleep(1 * time.Second)
							log(logz.LogInfo, "reconnecting... 4")
							break
						}
						log(logz.LogInfo, "func end")
					}
					log(logz.LogInfo, "func exit")
					// connectionChannel <- true
					msgCh <- nil
				}()

				log(logz.LogInfo, "connected")

				systray.SetIcon(iconGreen)

				for dbusMessage := range msgCh {
					if dbusMessage == nil {
						log(logz.LogInfo, "channel closed: exiting")
						connectionChannel <- true
						break
					}
					notiNotification, err := notilog.FromMessage(dbusMessage)
					if err != nil {
						if errors.Is(err, notilog.ErrNotANotification) {
							log(logz.LogDebug, "not a notification")
						} else {
							log(logz.LogWarn, "failed translating to message", err)
						}
						continue
					}
					log(logz.LogInfo, fmt.Sprintf("message intercepted: %v\n", notiNotification))

					// Converting notilog.Notification to our Notification because we have to send all strings.
					// It's easier to unmarshal strings on badge side.
					notification := Notification{
						Program:   notiNotification.Program,
						Title:     notiNotification.Title,
						Body:      notiNotification.Body,
						Sender:    notiNotification.Sender,
						Serial:    strconv.Itoa(int(notiNotification.Serial)),
						CreatedAt: notiNotification.CreatedAt.String(),
					}
					serialMessage, err := json.Marshal(notification)
					if err != nil {
						log(logz.LogErr, "failed to marshal notification", err)
						continue
					}

					// Send to serial
					if _, err = port.Write(serialMessage); err != nil {
						log(logz.LogErr, "failed to write to port", err)
						systray.SetIcon(iconRed)
						port.Close()
						time.Sleep(1 * time.Second)
						log(logz.LogInfo, "reconnecting... 5")
						connectionChannel <- true
						break
					}
				}
			}
		}
	}()
}

func addExitItem() {
	mExit := systray.AddMenuItem("Exit", "Exit the whole app")
	mExit.Enable()
	go func() {
		<-mExit.ClickedCh
		fmt.Println("Requesting exit")
		systray.Quit()
	}()
	systray.AddSeparator()
}

func onExit() {
	fmt.Println("onExit")
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
			// systray.Quit()
		}
	}
}
