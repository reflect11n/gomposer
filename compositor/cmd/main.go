package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var (
	rects   []xproto.Rectangle // Храним все прямоугольники
	rectsMu sync.Mutex         // Защищаем доступ к rects
)

func main() {
	os.Remove("/tmp/gocomp.sock")
	listener, err := net.Listen("unix", "/tmp/gocomp.sock")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer listener.Close()

	// X11 setup
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	winId, err := xproto.NewWindowId(X)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем окно с обработкой экспозиции
	err = xproto.CreateWindowChecked(
		X,
		screen.RootDepth,
		winId,
		screen.Root,
		0, 0, 500, 500, 0,
		xproto.WindowClassInputOutput,
		screen.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{
			0xFFFFFF, // Background: white
			xproto.EventMaskExposure,
		},
	).Check()
	if err != nil {
		log.Fatal(err)
	}

	xproto.MapWindow(X, winId)
	X.Sync() // Исправлено: Sync вместо Flush

	// Создаем общий GC
	gc, err := xproto.NewGcontextId(X)
	if err != nil {
		log.Fatal(err)
	}
	err = xproto.CreateGCChecked(X, gc, xproto.Drawable(winId), xproto.GcForeground, []uint32{0x000000}).Check()
	if err != nil {
		log.Fatal(err)
	}

	// Обработчик событий
	go func() {
		for {
			ev, err := X.WaitForEvent()
			if err != nil {
				log.Println("X11 error:", err)
				continue
			}
			switch ev {
			case xproto.ExposeEvent:
				redraw(X, winId, gc)
			}
		}
	}()

	fmt.Println("Composer: waiting for clients...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleClient(X, winId, gc, conn)
	}
}

func handleClient(X *xgb.Conn, win, gc xproto.Gcontext, conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 13)

	for {
		_, err := io.ReadFull(conn, buf) // Читаем ровно 13 байт
		if err != nil {
			return
		}

		if buf[0] == 0x01 { // drawRect
			x := binary.LittleEndian.Uint16(buf[2:])
			y := binary.LittleEndian.Uint16(buf[4:])
			w := binary.LittleEndian.Uint16(buf[6:])
			h := binary.LittleEndian.Uint16(buf[8:])
			r, g, b := buf[10], buf[11], buf[12]

			color := uint32(r)<<16 | uint32(g)<<8 | uint32(b)

			// Меняем цвет в существующем GC
			xproto.ChangeGC(X, gc, xproto.GcForeground, []uint32{color})

			rect := xproto.Rectangle{
				X:      int16(x),
				Y:      int16(y),
				Width:  w,
				Height: h,
			}

			// Сохраняем прямоугольник
			rectsMu.Lock()
			rects = append(rects, rect)
			rectsMu.Unlock()

			// Рисуем
			xproto.PolyFillRectangle(X, xproto.Drawable(win), gc, []xproto.Rectangle{rect})
			X.Sync() // Исправлено
		}
	}
}

// Перерисовка всего окна
func redraw(X *xgb.Conn, win xproto.Window, gc xproto.Gcontext) {
	// Очищаем окно
	bg := xproto.Rectangle{
		X:      0,
		Y:      0,
		Width:  500,
		Height: 500,
	}
	xproto.ChangeGC(X, gc, xproto.GcForeground, []uint32{0xFFFFFF})
	xproto.PolyFillRectangle(X, xproto.Drawable(win), gc, []xproto.Rectangle{bg})

	// Перерисовываем все прямоугольники
	rectsMu.Lock()
	defer rectsMu.Unlock()

	for _, rect := range rects {
		// Для демо - рисуем черным
		xproto.ChangeGC(X, gc, xproto.GcForeground, []uint32{0x000000})
		xproto.PolyFillRectangle(X, xproto.Drawable(win), gc, []xproto.Rectangle{rect})
	}
	X.Sync()
}
