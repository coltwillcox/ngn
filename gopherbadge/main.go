package main

import (
	"image/color"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"

	"tinygo.org/x/tinyfont/freemono"
)

const (
	carriageReturn = 10  // CarriageReturn represents the character code for a new line.
	newLine        = 13  // NewLine represents the character code for a new line.
	BufSize        = 128 // bufSize represents the buffer size of the input before flushing memory

	footerX           int16 = 8
	footerY           int16 = 217
	currentRectWidth  int16 = 8
	currentRectHeight int16 = 16
	currentRectSpace  int16 = 6
	maximumRects      int   = 10
)

var (
	uart = machine.Serial // uart is the name of the serial port stream

	display = st7789.New(machine.SPI0,
		machine.TFT_RST,       // TFT_RESET
		machine.TFT_WRX,       // TFT_DC
		machine.TFT_CS,        // TFT_CS
		machine.TFT_BACKLIGHT) // TFT_LITE

	// terminal represents the board terminal stream
	// terminal = tinyterm.NewTerminal(&display)

	// all the main colors in RGBA code.
	black  = color.RGBA{0, 0, 0, 255}
	white  = color.RGBA{255, 255, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	blue   = color.RGBA{0, 0, 255, 255}
	green  = color.RGBA{0, 255, 0, 255}
	violet = color.RGBA{116, 58, 213, 255}

	// font is the font used to display the text
	// font = &proggy.TinySZ8pt7b
	font = &freemono.Regular9pt7b
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display.Configure(st7789.Config{
		Rotation: st7789.ROTATION_270,
		Height:   320,
		Width:    240,
	})

	display.FillScreen(black)

	drawUI()
	printProgram("Program")
	printSender("Ime")
	printMessage("Poruka")
	drawFooter()

	// terminal.Configure(&tinyterm.Config{
	// 	Font:              font,
	// 	FontHeight:        24,
	// 	FontOffset:        6,
	// 	UseSoftwareScroll: true,
	// })

	// read the input
	// input := make([]byte, bufSize)
	// i := 0
	// tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 11, 11, "START", white)
	var i = int16(10)
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
		tinyfont.WriteLine(&display, font, 10, i, notification.Title, white)
		uart.Write([]byte("\r\n"))
		i = i + 20

		// if uart.Buffered() > 0 {
		// 	data, _ := uart.ReadByte()

		// 	// check the type of data
		// 	switch data {
		// 	// new line or new character, let's print the input and flush the memory
		// 	case CarriageReturn, NewLine:
		// 		writeInputAndFlush(input, i)
		// 		i = 0
		// 	// any other character, let's echo it
		// 	default:
		// 		// just echo the character
		// 		uart.WriteByte(data)
		// 		input[i] = data

		// 		tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 39, 39+20, string(data), white)

		// 		// _ = terminal.WriteByte(data)
		// 		i++

		// 		// out of range, let's write the input and flush
		// 		if i >= bufSize {
		// 			writeInputAndFlush(input, i)

		// 			i = 0
		// 		}
		// 	}
		// }
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

// writeInputAndFlush writes input on terminal and to serial port
// func writeInputAndFlush(input []byte, i int) {
// 	// write on the board
// 	// terminal.Write([]byte("\r\n"))
// 	// terminal.Write([]byte("You typed: "))
// 	// terminal.Write(input[:i])
// 	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 39, 39+20, string(input), white)

// 	// write on the serial port
// 	// uart.Write([]byte("\r\n"))
// 	// uart.Write([]byte("You typed: "))

// 	// uart.Write(input[:i])

// 	// write new lines in terminal and to serial port
// 	// terminal.Write([]byte("\r\n"))

// 	uart.Write([]byte("\r\n"))
// }

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
	tinydraw.Rectangle(&display, 0, 0, 320, 240, violet)
	tinydraw.Rectangle(&display, 8, 8, 264, 30, violet)
	tinydraw.Rectangle(&display, 8, 48, 304, 30, violet)
	tinydraw.Rectangle(&display, 8, 88, 304, 122, violet)
}

func printProgram(program string) {
	tinyfont.WriteLine(&display, font, 13, 27, program, white)
}

func printSender(sender string) {
	tinyfont.WriteLine(&display, font, 13, 67, sender, white)
}

func printMessage(message string) {
	tinyfont.WriteLine(&display, font, 13, 107, message, white)
}

func drawFooter() {
	tinyfont.WriteLine(&display, font, footerX+225, footerY+13, "L/R/A/B", violet)
	for i := 0; i < maximumRects; i++ {
		tinydraw.Rectangle(&display, footerX+(int16(i)*(currentRectWidth+currentRectSpace)), footerY, currentRectWidth, currentRectHeight, violet)
	}
}
