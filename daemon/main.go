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
	badgePort     = "/dev/ttyACM0"
	connectCheck  = 5 // Seconds
	messageLength = 128
)

// Icons taken from https://github.com/egonelbre/gophers
var (
	//go:embed icons/online.png
	iconOnline []byte
	//go:embed icons/offline.png
	iconOffline []byte
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Neon Gopher Notifications")
	systray.SetIcon(iconOffline)
	systray.SetTooltip("Connecting...")
	addExitItem()

	go func() {
		log := logFn()

		msgCh := make(chan *dbus.Message, 100)
		c := conductor.Simple[notilog.Action]()

		l, err := notilog.NewNotiListener(msgCh)
		if err != nil {
			log(logz.LogErr, "failed to initialize listener", err)
			systray.Quit()
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
				port, err := serial.Open(badgePort, &serial.Mode{})
				if err != nil {
					prepareForReconnect(log, port, "failed to open port", err)
					connectionChannel <- true
					break
				}

				// Port existing/connected listener.
				go func() {
					for {
						time.Sleep(connectCheck * time.Second)
						portsNames, err := serial.GetPortsList()
						if err != nil {
							prepareForReconnect(log, port, "failed to get ports", err)
							break
						}
						if len(portsNames) == 0 {
							prepareForReconnect(log, port, "no serial ports found", nil)
							break
						}
						existing := false
						for _, portName := range portsNames {
							if portName == badgePort {
								existing = true
								break
							}
						}
						if !existing {
							prepareForReconnect(log, port, "port does not exist", nil)
							break
						}
					}
					msgCh <- nil
				}()

				log(logz.LogInfo, "connected")

				systray.SetIcon(iconOnline)
				systray.SetTooltip("Connected")

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
					// If message is larger than that, end will be chopped, therefore, we are spliting message into chunks of 128 bytes.
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
						fmt.Println(string(serialMessagePart))
						if _, err = port.Write(serialMessagePart); err != nil {
							prepareForReconnect(log, port, "failed to write to port", err)
							connectionChannel <- true
							break
						}
						// Give some time to Gopher Badge to process each part.
						// Required for multipart messages.
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}
	}()
}

func prepareForReconnect(log func(logz.LogLevel, string, ...error), port serial.Port, message string, err error) {
	log(logz.LogErr, message, err)
	systray.SetIcon(iconOffline)
	systray.SetTooltip("Connecting...")
	if port != nil {
		port.Close()

	}
	log(logz.LogInfo, "connecting...")
	time.Sleep(connectCheck * time.Second)
}

func onExit() {
	fmt.Println("onExit")
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

func logFn() func(logz.LogLevel, string, ...error) {
	log := zlog.NewConsoleLogger()
	return func(level logz.LogLevel, msg string, errs ...error) {
		data := map[string]any{"msg": msg}
		if len(errs) > 0 && errs[0] != nil {
			data["err"] = errs[0].Error()
		}
		log.Log(level, data)
		// if level == logz.LogErr {
		// 	systray.Quit()
		// }
	}
}
