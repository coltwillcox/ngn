// TODO Menu option to pause.
// TODO Change icon on incoming notification.
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
	badgePort        = "/dev/ttyACM0"
	commandClear     = "clear"
	messageLength    = 128
	timeConnectCheck = 5   // Seconds.
	timeRest         = 100 // Milliseconds.
)

// Icons taken from https://github.com/egonelbre/gophers
var (
	//go:embed icons/online.png
	iconOnline []byte
	//go:embed icons/offline.png
	iconOffline []byte

	channelConnection chan bool
	channelMessage    chan *dbus.Message
	log               func(logz.LogLevel, string, ...error)
)

func main() {
	initialize()
	systray.Run(onReady, onExit)
}

func initialize() {
	channelMessage = make(chan *dbus.Message, 100)
	channelConnection = make(chan bool, 1)
	log = logFn()
}

func onReady() {
	systray.SetIcon(iconOffline)
	systray.SetTitle("Neon Gopher Notifications")
	systray.SetTooltip("Connecting...")
	time.Sleep(timeRest * time.Millisecond) // Give some time to set icon.

	mClear := addClearItem()
	addPauseItem()
	mExit := addExitItem()

	go func() {
		var err error
		if listener, err := notilog.NewNotiListener(channelMessage); err != nil {
			log(logz.LogErr, "failed to initialize listener", err)
			systray.Quit()
		} else {
			go func() {
				if err := listener.Run(conductor.Simple[notilog.Action]()); err != nil {
					log(logz.LogErr, "execution failed", err)
				}
				log(logz.LogInfo, "exited successfully")
			}()
		}

		var port serial.Port
		channelConnection <- true
		for {
			select {
			case <-mClear.ClickedCh:
				if _, err = port.Write([]byte(commandClear)); err != nil {
					prepareForReconnect(log, &port, "failed to write to port", err)
					channelConnection <- true
					continue
				}
				// Give some time to Gopher Badge to process each part. Required for multipart messages.
				time.Sleep(timeRest * time.Millisecond)
			case <-mExit.ClickedCh:
				systray.Quit()
			case dbusMessage := <-channelMessage:
				if port == nil || dbusMessage == nil {
					continue
				}

				notiNotification, err := notilog.FromMessage(dbusMessage)
				if err != nil {
					if errors.Is(err, notilog.ErrNotANotification) {
						log(logz.LogDebug, "message not a notification")
					} else {
						log(logz.LogWarn, "failed translating to message", err)
					}
					continue
				}
				log(logz.LogInfo, fmt.Sprintf("message intercepted: %v\n", notiNotification))

				// Converting notilog.Notification to our Notification because we have to send all types as strings.
				// It's easier to unmarshal strings on badge side.
				notification := Notification{
					Program:   notiNotification.Program,
					Title:     notiNotification.Title,
					Body:      notiNotification.Body,
					Sender:    notiNotification.Sender,
					Serial:    strconv.Itoa(int(notiNotification.Serial)),
					CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				}
				serialMessage, err := json.Marshal(notification)
				if err != nil {
					log(logz.LogErr, "failed to marshal notification", err)
					continue
				}

				// Maximum single message length that can be transmitted to Gopher Badge is 128 bytes.
				// If message is larger than that, end will be truncated, therefore, we are spliting message into chunks of 128 bytes.
				serialMessageParts := [][]byte{}
				for i := 0; i < len(serialMessage); i += messageLength {
					end := i + messageLength

					if end > len(serialMessage) {
						end = len(serialMessage)
					}

					serialMessageParts = append(serialMessageParts, serialMessage[i:end])
				}

				// Send to serial.
				for _, serialMessagePart := range serialMessageParts {
					if _, err = port.Write(serialMessagePart); err != nil {
						log(logz.LogErr, "failed to write to port", err)
						break
					}
					// Give some time to Gopher Badge to process each part. Required for multipart messages.
					time.Sleep(timeRest * time.Millisecond)
				}
			case <-channelConnection:
				port, err = serial.Open(badgePort, &serial.Mode{})
				if err != nil {
					prepareForReconnect(log, &port, "failed to open port", err)
					channelConnection <- true
					continue
				}

				systray.SetIcon(iconOnline)
				systray.SetTooltip("Connected")
				log(logz.LogInfo, "connected")

				// Port existing/connected listener.
				go func() {
					for {
						time.Sleep(timeConnectCheck * time.Second)
						portsNames, err := serial.GetPortsList()
						if err != nil {
							prepareForReconnect(log, &port, "failed to get ports", err)
							break
						}
						if len(portsNames) == 0 {
							prepareForReconnect(log, &port, "no serial ports found", nil)
							break
						}
						existing := false
						for _, portName := range portsNames {
							if portName == badgePort {
								existing = true
								continue
							}
						}
						if !existing {
							prepareForReconnect(log, &port, "port does not exist", nil)
							break
						}
					}
					channelConnection <- true
				}()
			}
		}
	}()
}

func onExit() {
	log(logz.LogInfo, "exiting...")
}

func prepareForReconnect(log func(logz.LogLevel, string, ...error), port *serial.Port, message string, err error) {
	log(logz.LogErr, message, err)
	systray.SetIcon(iconOffline)
	systray.SetTooltip("Reconnecting...")
	if *port != nil {
		(*port).Close()
		*port = nil
	}
	log(logz.LogInfo, "reconnecting...")
	time.Sleep(timeConnectCheck * time.Second)
}

func addClearItem() *systray.MenuItem {
	mClear := systray.AddMenuItem("Clear", "Clear history")
	mClear.Enable()
	return mClear
}

func addPauseItem() *systray.MenuItem {
	mPause := systray.AddMenuItemCheckbox("Pause", "Pause transmitting", true)
	mPause.Enable()
	return mPause
}

func addExitItem() *systray.MenuItem {
	mExit := systray.AddMenuItem("Exit", "Exit the whole app")
	mExit.Enable()
	return mExit
}

func logFn() func(logz.LogLevel, string, ...error) {
	log := zlog.NewConsoleLogger()
	return func(level logz.LogLevel, msg string, errs ...error) {
		data := map[string]any{"msg": msg}
		if len(errs) > 0 && errs[0] != nil {
			data["err"] = errs[0].Error()
		}
		log.Log(level, data)
	}
}
