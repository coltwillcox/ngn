// TODO Deleting single notification should not stay on empty field.
// TODO Paused icon.
package main

import (
	"image/color"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/drivers/ws2812"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"

	"github.com/coltwillcox/ngn/gopherbadge/views"
)

type Notification struct {
	Program   string
	Title     string
	Body      string
	Sender    string
	Serial    string
	CreatedAt string
	Icon      string
}

const (
	carriageReturn        = 10  // Character code for a new line.
	newLine               = 13  // Character code for a new line.
	timeRest              = 10  // Milliseconds.
	timeDimmer            = 100 // Milliseconds.
	maximumRects   int    = 10
	historySize    int    = 10
	footerX        int16  = 0
	footerY        int16  = 217
	pageRectWidth  int16  = 8
	pageRectHeight int16  = 16
	pageRectSpace  int16  = 6
	screenWidth    int16  = 320
	screenHeight   int16  = 240
	textViewHeight int16  = 30
	margin         int16  = 8
	commandClear   string = "clear"
	separator      rune   = '*'
)

var (
	uart                 = machine.Serial // Serial port stream.
	display              = st7789.New(machine.SPI0, machine.TFT_RST, machine.TFT_WRX, machine.TFT_CS, machine.TFT_BACKLIGHT)
	leds                 = machine.NEOPIXELS
	ledsDriver           = ws2812.New(leds)
	black                = color.RGBA{0, 0, 0, 255}
	white                = color.RGBA{255, 255, 255, 255}
	red                  = color.RGBA{255, 0, 0, 255}
	blue                 = color.RGBA{0, 0, 255, 255}
	green                = color.RGBA{0, 255, 0, 255}
	violet               = color.RGBA{116, 58, 213, 255}
	yellow               = color.RGBA{255, 255, 0, 255}
	font                 = &freemono.Regular9pt7b // Font used to display the text.
	screenBorderRectView = views.RectView{}
	programTextView      = views.TextView{}
	timeTextView         = views.TextView{}
	messageTextView      = views.TextView{}
	iconImageView        = views.ImageView{}
	history              = make([]Notification, 0, historySize)
	pagesRectViews       = make([]views.RectView, historySize)
	currentPage          = 0
	buttonA              = machine.BUTTON_A
	buttonB              = machine.BUTTON_B
	buttonUp             = machine.BUTTON_UP
	buttonLeft           = machine.BUTTON_LEFT
	buttonDown           = machine.BUTTON_DOWN
	buttonRight          = machine.BUTTON_RIGHT
	ledOpacity           = -1
)

func main() {
	configure()
	drawUI()
	drawFooter()

	channelMessage := make(chan string, 1)

	go func() {
		for {
			time.Sleep(timeDimmer * time.Millisecond)
			dimLeds()
			checkButtons()
		}
	}()

	go func() {
		text := make([]byte, 0)
		for {
			time.Sleep(timeRest * time.Millisecond)
			if uart.Buffered() == 0 {
				continue
			}

			for {
				singleByte, _ := uart.ReadByte()
				switch singleByte {
				case carriageReturn, newLine:
					continue
				case byte(separator):
					channelMessage <- string(text)
					text = nil
					break
				default:
					text = append(text, singleByte)
				}

				if uart.Buffered() == 0 {
					break
				}
			}
		}
	}()

	for {
		select {
		case message := <-channelMessage:
			switch message {
			case commandClear:
				clearHistory()
			default:
				addToHistory(fromMessage(message))
				drawCurrentPage()
				drawFooter()
				lightUpLeds()
			}
		}
	}
}

func configure() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	leds.Configure(machine.PinConfig{Mode: machine.PinOutput})

	buttonA.Configure(machine.PinConfig{Mode: machine.PinInput})
	buttonB.Configure(machine.PinConfig{Mode: machine.PinInput})
	buttonUp.Configure(machine.PinConfig{Mode: machine.PinInput})
	buttonLeft.Configure(machine.PinConfig{Mode: machine.PinInput})
	buttonDown.Configure(machine.PinConfig{Mode: machine.PinInput})
	buttonRight.Configure(machine.PinConfig{Mode: machine.PinInput})

	display.Configure(st7789.Config{
		Rotation: st7789.ROTATION_270,
		Height:   screenWidth,
		Width:    screenHeight,
	})

	for i := 0; i < len(pagesRectViews); i++ {
		pagesRectViews[i].SetDisplay(&display).SetDimensions(footerX+margin+(int16(i)*(pageRectWidth+pageRectSpace)), footerY, pageRectWidth, pageRectHeight).SetColor(&violet)
	}
}

func drawUI() {
	screenBorderRectView.SetDisplay(&display).SetColor(&violet).SetDimensions(0, 0, screenWidth, screenHeight).Draw()
	programTextView.SetDisplay(&display).SetFont(font).SetFontColor(&yellow).SetColor(&violet).SetDimensions(margin, margin, screenWidth-margin*2-40, textViewHeight).Draw()
	timeTextView.SetDisplay(&display).SetFont(font).SetFontColor(&yellow).SetColor(&violet).SetDimensions(margin, textViewHeight+margin*2-1, screenWidth-margin*2, textViewHeight).Draw()
	messageTextView.SetDisplay(&display).SetFont(font).SetFontColor(&yellow).SetColor(&violet).SetDimensions(margin, textViewHeight*2+margin*3-2, 304, 126).Draw()
	iconImageView.SetDisplay(&display).SetBackgroundColor(&black).SetDimensions(281, margin, textViewHeight, textViewHeight).Draw()
}

func drawFooter() {
	for i := 0; i < historySize; i++ {
		color := violet
		backgroundColor := black
		if i == currentPage {
			color = yellow
		}
		if len(history) > i {
			backgroundColor = violet
		}
		pagesRectViews[i].SetColor(&color).SetBackgroundColor(&backgroundColor).Draw()
	}
	tinyfont.WriteLine(&display, font, footerX+margin+225, footerY+13, "L/R/A/B", violet)
}

func fromMessage(message string) Notification {
	notification := Notification{}
	messageTrimmed := strings.TrimSuffix(strings.TrimPrefix(message, "{\""), "\"}")
	parts := strings.Split(messageTrimmed, "\",\"")
	for _, part := range parts {
		keyValue := strings.Split(part, "\":\"")
		switch keyValue[0] {
		case "program":
			notification.Program = keyValue[1]
		case "title":
			notification.Title = keyValue[1]
		case "body":
			notification.Body = keyValue[1]
		case "sender":
			notification.Sender = keyValue[1]
		case "serial":
			notification.Serial = keyValue[1]
		case "created_at":
			notification.CreatedAt = keyValue[1]
		case "icon":
			notification.Icon = keyValue[1]
		}
	}

	return notification
}

func addToHistory(notification Notification) {
	if len(history) >= historySize {
		history = history[1:]
	}
	history = append(history, notification)
	currentPage = len(history) - 1
}

func removeFromHistory(i int) bool {
	if len(history) == 0 || len(history) <= i {
		return false
	}

	history = append(history[:i], history[i+1:]...)
	return true
}

func clearHistory() {
	if len(history) != 0 {
		history = make([]Notification, 0, historySize)
		currentPage = 0
		drawCurrentPage()
		drawFooter()
	}
	shutDownLeds()
}

func chunks(text string, chunkSize int, maximumChunks int) []string {
	if len(text) == 0 {
		return nil
	}
	if chunkSize >= len(text) {
		return []string{text}
	}
	var chunks []string = make([]string, 0, maximumChunks)
	currentLength := 0
	currentStart := 0
	for i := range text {
		if currentLength == chunkSize {
			chunks = append(chunks, text[currentStart:i])
			currentLength = 0
			currentStart = i
		}
		currentLength++
		if len(chunks) >= maximumChunks-1 {
			break
		}
	}
	chunks = append(chunks, text[currentStart:])
	return chunks
}

func drawCurrentPage() {
	if len(history)-1 < currentPage || len(history) == 0 {
		programTextView.SetText("")
		timeTextView.SetText("")
		messageTextView.SetText("")
		iconImageView.SetImage("")
		return
	}

	currentNotification := history[currentPage]
	programTextView.SetText(currentNotification.Program)
	timeTextView.SetText(currentNotification.CreatedAt)
	messageTextView.SetText(currentNotification.Title)
	iconImageView.SetImage(currentNotification.Icon)
}

func checkButtons() {
	if !buttonLeft.Get() {
		navigatePage(false)
	} else if !buttonRight.Get() {
		navigatePage(true)
	} else if !buttonB.Get() {
		if removeFromHistory(currentPage) {
			drawCurrentPage()
			drawFooter()
		}
		shutDownLeds()
	} else if !buttonA.Get() {
		clearHistory()
	}
}

func navigatePage(advance bool) {
	move := 1
	if !advance {
		move = -1
	}
	currentPage += move
	if currentPage < 0 {
		currentPage = 0
	} else if currentPage > historySize-1 {
		currentPage = historySize - 1
	}
	drawCurrentPage()
	drawFooter()
}

func dimLeds() {
	if ledOpacity <= 0 {
		return
	}

	ledsDriver.WriteColors([]color.RGBA{color.RGBA{uint8(ledOpacity), 0, 0, 255}, color.RGBA{0, 0, uint8(ledOpacity), 255}})
	ledOpacity -= 10
}

func lightUpLeds() {
	ledOpacity = 255
}

func shutDownLeds() {
	if len(history) != 0 {
		return
	}

	ledOpacity = 0
	ledsDriver.WriteColors([]color.RGBA{color.RGBA{uint8(ledOpacity), 0, 0, 255}, color.RGBA{0, 0, uint8(ledOpacity), 255}})
}
