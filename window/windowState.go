package window

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const (
	_NET_WM_STATE       = "_NET_WM_STATE"
	_NET_WM_STATE_ABOVE = "_NET_WM_STATE_ABOVE"
)

func SetWindowAlwaysOnTop(win fyne.Window) {
	nativeWin, ok := win.(driver.NativeWindow)
	if !ok {
		panic("this will never happen for a top-level window")
	}

	nativeWin.RunNative(func(ctx any) {
		switch runtime.GOOS {
		case "linux":
			x11Context := ctx.(driver.X11WindowContext)
			x11Window := x11Context.WindowHandle

			// Connect to the X server
			x, err := xgb.NewConn()
			if err != nil {
				log.Fatalf("Failed to connect to X server: %v", err)
			}
			defer x.Close()

			// Intern atoms
			netWmStateAtom, err := xproto.InternAtom(x, true, uint16(len(_NET_WM_STATE)), _NET_WM_STATE).
				Reply()
			if err != nil {
				log.Fatalf("Failed to intern atom %s: %v", _NET_WM_STATE, err)
			}

			netWmStateAboveAtom, err := xproto.InternAtom(x, true, uint16(len(_NET_WM_STATE_ABOVE)), _NET_WM_STATE_ABOVE).
				Reply()
			if err != nil {
				log.Fatalf("Failed to intern atom %s: %v", _NET_WM_STATE_ABOVE, err)
			}

			// Get the root window
			setup := xproto.Setup(x)
			root := setup.DefaultScreen(x).Root

			// Prepare client message
			data := []uint32{
				1, // _NET_WM_STATE_ADD
				uint32(netWmStateAboveAtom.Atom),
				0,
				0,
				0,
			}
			event := xproto.ClientMessageEvent{
				Format: 32,
				Window: xproto.Window(x11Window),
				Type:   netWmStateAtom.Atom,
				Data:   xproto.ClientMessageDataUnionData32New(data),
			}

			// Send the client message
			err = xproto.SendEventChecked(x, false, root, xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify, string(event.Bytes())).
				Check()
			if err != nil {
				log.Fatalf("Failed to send _NET_WM_STATE client message: %v", err)
			}
		}
	})
}

func SetWindowOnTopRightCorner(win fyne.Window) {
	nativeWin, ok := win.(driver.NativeWindow)
	if !ok {
		panic("this will never happen for a top-level window")
	}

	nativeWin.RunNative(func(ctx any) {
		switch runtime.GOOS {
		case "linux":
			x11Context := ctx.(driver.X11WindowContext)
			x11Window := x11Context.WindowHandle

			// Connect to the X server
			x, err := xgb.NewConn()
			if err != nil {
				log.Fatalf("Failed to connect to X server: %v", err)
			}
			defer x.Close()

			// Get screen size
			setup := xproto.Setup(x)
			screen := setup.DefaultScreen(x)

			// Calculate top-right position
			width := int16(screen.WidthInPixels) - 320 // Screen width minus window width
			height := int16(0)                         // Top position

			// Move the window to the top-right corner
			err = xproto.ConfigureWindowChecked(x, xproto.Window(x11Window), xproto.ConfigWindowX|xproto.ConfigWindowY, []uint32{uint32(width), uint32(height)}).Check()
			if err != nil {
				log.Fatalf("Failed to move window: %v", err)
			}
		}
	})
}
