package notilog

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	memberNotification = "\"Notify\""
)

var (
	ErrMalformedMsg     = errors.New("malformed message")
	ErrNotANotification = errors.New("message is not a notification")
)

type Notification struct {
	Program   string    `json:"program"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Sender    string    `json:"sender"`
	Serial    uint32    `json:"serial"`
	CreatedAt time.Time `json:"created_at"`
}

func (n *Notification) GetField(field string) (string, bool) {
	switch field {
	case "program":
		return n.Program, true
	case "title":
		return n.Title, true
	case "body":
		return n.Body, true
	case "sender":
		return n.Sender, true
	case "serial":
		return fmt.Sprint(n.Serial), true
	case "created_at":
		return n.CreatedAt.Format(time.RFC3339), true
	}
	return "", false
}

func (n *Notification) WithEscapedBody() *Notification {
	return &Notification{
		Program:   n.Program,
		Title:     n.Title,
		Sender:    n.Sender,
		Serial:    n.Serial,
		CreatedAt: n.CreatedAt,
		Body:      strings.Replace(n.Body, "\n", " ⤶ ", -1),
	}
}

func FromMessage(msg *dbus.Message) (*Notification, error) {
	if msg == nil {
		return nil, ErrMalformedMsg
	}

	if member, ok := msg.Headers[dbus.FieldMember]; !ok || member.String() != memberNotification {
		return nil, ErrNotANotification
	}

	if len(msg.Body) != 8 {
		return nil, ErrMalformedMsg
	}

	sender, ok := msg.Headers[dbus.FieldSender]
	if !ok {
		return nil, ErrMalformedMsg
	}

	program, ok := msg.Body[0].(string)
	if !ok {
		return nil, ErrMalformedMsg
	}

	serial, ok := msg.Body[1].(uint32)
	if !ok {
		return nil, ErrMalformedMsg
	}

	title, ok := msg.Body[3].(string)
	if !ok {
		return nil, ErrMalformedMsg
	}

	body, ok := msg.Body[4].(string)
	if !ok {
		return nil, ErrMalformedMsg
	}

	return &Notification{
		Program:   program,
		Serial:    serial,
		Title:     title,
		Body:      body,
		Sender:    sender.String(),
		CreatedAt: time.Now(),
	}, nil
}

func TranslateHeaders(headerMap map[dbus.HeaderField]dbus.Variant) map[string]string {
	res := make(map[string]string)

	for k, v := range headerMap {
		switch k {
		case dbus.FieldPath:
			res["path"] = v.String()
		case dbus.FieldInterface:
			res["interface"] = v.String()
		case dbus.FieldMember:
			res["member"] = v.String()
		case dbus.FieldErrorName:
			res["error_name"] = v.String()
		case dbus.FieldReplySerial:
			res["reply_serial"] = v.String()
		case dbus.FieldDestination:
			res["destination"] = v.String()
		case dbus.FieldSender:
			res["sender"] = v.String()
		case dbus.FieldSignature:
			res["signature"] = v.String()
		case dbus.FieldUnixFDs:
			res["unix_fds"] = v.String()
		default:
			panic("unexpected")
		}
	}

	return res
}
