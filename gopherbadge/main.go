package main

import (
	"image/color"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"

	"github.com/coltwillcox/ngn/gopherbadge/views"
)

const (
	carriageReturn          = 10 // Character code for a new line.
	newLine                 = 13 // Character code for a new line.
	footerX           int16 = 0
	footerY           int16 = 217
	currentRectWidth  int16 = 8
	currentRectHeight int16 = 16
	currentRectSpace  int16 = 6
	maximumRects      int   = 10
	screenWidth       int16 = 320
	screenHeight      int16 = 240
	margin            int16 = 8
	textViewHeight    int16 = 30
)

var (
	uart = machine.Serial // Serial port stream.

	display = st7789.New(machine.SPI0,
		machine.TFT_RST,       // TFT_RESET
		machine.TFT_WRX,       // TFT_DC
		machine.TFT_CS,        // TFT_CS
		machine.TFT_BACKLIGHT) // TFT_LITE

	// Main colors in RGBA code.
	black  = color.RGBA{0, 0, 0, 255}
	white  = color.RGBA{255, 255, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	blue   = color.RGBA{0, 0, 255, 255}
	green  = color.RGBA{0, 255, 0, 255}
	violet = color.RGBA{116, 58, 213, 255}

	font        = &freemono.Regular9pt7b // Font used to display the text.
	historySize = 10
	history     = make([]*Notification, 0, historySize)
	currentPage = 0

	screenBorderRectView = views.RectView{}
	programTextView      = views.TextView{}
	senderTextView       = views.TextView{}
	messageTextView      = views.TextView{}
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display.Configure(st7789.Config{
		Rotation: st7789.ROTATION_270,
		Height:   screenWidth,
		Width:    screenHeight,
	})

	display.FillScreen(black)
	drawUI()
	drawFooter()

	// printProgram("123456789012345678901234567890")
	// printSender("123456789012345678901234567890")
	// printMessage("123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")

	for {
		if uart.Buffered() == 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		serialMessage := make([]byte, 0)

		for uart.Buffered() > 0 {
			singleByte, _ := uart.ReadByte()
			switch singleByte {
			case carriageReturn, newLine:
				continue
			default:
				serialMessage = append(serialMessage, singleByte)
			}
		}

		notification := fromMessage(serialMessage)
		addToHistory(notification)
		drawCurrentPage()
		drawFooter()
		uart.Write([]byte("\r\n"))
	}
}

func fromMessage(serialMessage []byte) *Notification {
	notification := Notification{}
	message := string(serialMessage)
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
		}
	}

	return &notification
}

// TODO Separate fields into eg 2 parts, so they can be wrapped
type Notification struct {
	Program string
	Title   string
	Body    string
	Sender  string
	Serial  string
	// CreatedAt time.Time `json:"created_at"`
}

func drawUI() {
	screenBorderRectView.SetDisplay(&display).SetColor(&violet).SetDimensions(0, 0, screenWidth, screenHeight).Draw()
	programTextView.SetDisplay(&display).SetFont(font).SetColor(&violet).SetDimensions(margin, margin, screenWidth-margin*2-40, textViewHeight).SetText("111111111111111112222222222222222222222").Draw()
	senderTextView.SetDisplay(&display).SetFont(font).SetColor(&violet).SetDimensions(margin, textViewHeight+margin*2-1, screenWidth-margin*2, textViewHeight).Draw().SetText("111111111111111112222222222222222222222")
	messageTextView.SetDisplay(&display).SetFont(font).SetColor(&violet).SetDimensions(margin, textViewHeight*2+margin*3-2, 304, 126).Draw().SetText("1111111111aaaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaaaaaaaa11112222222222222222222222w12345678901111111111aaaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaaaaa1aaaaaaaaaaaaaaaaaaaaaaaaa11112222222222222222222222w1234567890")
}

func drawFooter() {
	// tinydraw.FilledRectangle(&display, footerX, footerY, screenWidth, screenHeight-footerY, black)
	for i := 0; i < historySize; i++ {
		if len(history) > i && history[i] != nil {
			drawFilledRectangle(&display, footerX+margin+(int16(i)*(currentRectWidth+currentRectSpace)), footerY, currentRectWidth, currentRectHeight, violet)
		} else {
			drawRectangle(&display, footerX+margin+(int16(i)*(currentRectWidth+currentRectSpace)), footerY, currentRectWidth, currentRectHeight, violet)
		}
	}
	tinyfont.WriteLine(&display, font, footerX+margin+225, footerY+13, "L/R/A/B", violet)
}

func addToHistory(notification *Notification) {
	if len(history) >= historySize {
		history = history[1:]
	}
	history = append(history, notification)
	currentPage = len(history) - 1
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
	if len(history)-1 > currentPage {
		return
	}

	currentNotification := history[currentPage]
	if currentNotification == nil {
		return
	}

	programTextView.SetText(currentNotification.Program)
	senderTextView.SetText(currentNotification.Sender)
	messageTextView.SetText(currentNotification.Title)
}

func drawRectangle(displayer *st7789.Device, x int16, y int16, w int16, h int16, color color.RGBA) {
	displayer.DrawFastHLine(x, x+w-1, y, color)
	displayer.DrawFastHLine(x, x+w-1, y+h-1, color)
	displayer.DrawFastVLine(x, y, y+h-1, color)
	displayer.DrawFastVLine(x+w-1, y, y+h-1, color)
}

func drawFilledRectangle(displayer *st7789.Device, x int16, y int16, w int16, h int16, color color.RGBA) {
	displayer.FillRectangle(x, y, w, h, color)
}
