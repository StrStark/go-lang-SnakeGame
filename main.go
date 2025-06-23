package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

var Margin int = 3
var X int
var Y int
var FrootX int
var FrootY int
var HeadX int
var HeadY int
var CurrentScore int
var hasEaten bool = false
var bodyX = []int{}
var bodyY = []int{}

type Direction int

const (
	Updir Direction = iota
	Rightdir
	Downdir
	Leftdir
)

var direction Direction = Downdir

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")
)

const (
	ENABLE_PROCESSED_INPUT = 0x0001
	ENABLE_LINE_INPUT      = 0x0002
	ENABLE_ECHO_INPUT      = 0x0004

	KEY_EVENT = 0x0001
)

type inputRecord struct {
	EventType uint16
	_         uint16
	Event     keyEventRecord
}

type keyEventRecord struct {
	bKeyDown          int32
	wRepeatCount      uint16
	wVirtualKeyCode   uint16
	wVirtualScanCode  uint16
	unicodeChar       uint16
	dwControlKeyState uint32
}

func getConsoleMode(handle syscall.Handle) (mode uint32, err error) {
	r1, _, e1 := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return 0, e1
	}
	return mode, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	r1, _, e1 := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if r1 == 0 {
		return e1
	}
	return nil
}
func listenToInput(handle syscall.Handle) {
	var record inputRecord
	var read uint32

	for {
		r1, _, err := procReadConsoleInput.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&record)),
			1,
			uintptr(unsafe.Pointer(&read)),
		)
		if r1 == 0 {
			fmt.Println("Read error:", err)
			return
		}

		if record.EventType == KEY_EVENT && record.Event.bKeyDown != 0 {
			switch record.Event.wVirtualKeyCode {
			case 0x26: // Up
				if direction != Downdir {
					direction = Updir
				}
			case 0x28: // Down
				if direction != Updir {
					direction = Downdir
				}
			case 0x27: // Right
				if direction != Leftdir {
					direction = Rightdir
				}
			case 0x25: // Left
				if direction != Rightdir {
					direction = Leftdir
				}
			case 'Q', 'q':
				fmt.Println("\nQuitting...")
				os.Exit(0)
			}
		}
	}
}

func main() {
	handle := syscall.Handle(os.Stdin.Fd())

	originalMode, err := getConsoleMode(handle)
	if err != nil {
		fmt.Println("Error getting console mode:", err)
		return
	}
	rawMode := originalMode &^ (ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT)
	if err := setConsoleMode(handle, rawMode); err != nil {
		fmt.Println("Error setting console mode:", err)
		return
	}
	defer setConsoleMode(handle, originalMode)

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		setConsoleMode(handle, originalMode)
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()
	Setup()
	go listenToInput(handle)

	ticker := time.NewTicker(20 * time.Millisecond) // Snake speed
	defer ticker.Stop()
	for range ticker.C {

		ClearScreen()

		if HeadX == FrootX && HeadY == FrootY {
			FrootX, FrootY = FrootGenerator()
			CurrentScore += 1
			addBody()
		} else if CurrentScore > 0 {

			UpdateBodyLocations()

		}
		Move()
		createBoard()
		if checkLost() {
			break
		}

	}
	print("Tnanks for playing the game \n contact : rezafathisamani1383@gmail.com")
}

func checkLost() bool {
	for i := 0; i < CurrentScore; i++ {
		if HeadX == bodyX[i] && HeadY == bodyY[i] {
			return true
		}
	}
	return false
}

func Move() {
	switch direction {
	case Updir:
		if HeadX == 1 {
			HeadX = X - 1
		} else {
			HeadX = (HeadX - 1) % (X - 1)
		}
	case Downdir:
		HeadX = (HeadX + 1) % X
	case Rightdir:
		HeadY = (HeadY + 1) % Y
	case Leftdir:
		if HeadY == 1 {
			HeadY = Y - 1
		} else {
			HeadY = (HeadY - 1) % (Y - 1)
		}
	}
}

func addBody() {
	bodyX = append(bodyX, HeadX)
	bodyY = append(bodyY, HeadY)
	hasEaten = true

}

func Setup() {
	X = 30
	Y = 60
	HeadX = rand.Intn(X-2) + 1
	HeadY = rand.Intn(Y-2) + 1
	FrootX, FrootY = FrootGenerator()
	CurrentScore = 0
	bodyX = []int{}
	bodyY = []int{}
}

func ClearScreen() {
	fmt.Print("\033[H")

}

func createBoard() {
	var isbody bool
	var isdot = false
	for i := 0; i < Margin*3; i++ {
		print("\n")
	}
	for i := 0; i < X; i++ {
		for i := 0; i < Margin; i++ {
			print("\t")
		}
		for j := 0; j < Y; j++ {

			isbody = false
			for k := 0; k < CurrentScore; k++ {
				if i == bodyX[k] && j == bodyY[k] {
					isbody = true
				}
			}
			if isbody {
				if isdot {
					print(".")
					isdot = false
				} else {
					print("o")
					isdot = true
				}
			} else if i == 0 || i == X-1 || j == 0 || j == Y-1 {
				print("#")
			} else if i == HeadX && j == HeadY {
				print("O")
			} else if i == FrootX && j == FrootY {
				print("F")
			} else {
				print(" ")
			}
		}
		println()

	}
	for i := 0; i < Margin*3; i++ {
		print("\n")
	}
}

func FrootGenerator() (int, int) {
	var x = rand.Intn(X-2) + 1
	var y = rand.Intn(Y-2) + 1
	return x, y
}

func UpdateBodyLocations() {
	for i := 0; i < CurrentScore-1; i++ {
		bodyX[i] = bodyX[i+1]
		bodyY[i] = bodyY[i+1]
	}
	bodyX[CurrentScore-1] = HeadX
	bodyY[CurrentScore-1] = HeadY
}
