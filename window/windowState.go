package window

import (
	"log"
	"runtime"
	"stinote/consts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const (
	_NET_WM_STATE       = "_NET_WM_STATE"
	_NET_WM_STATE_ABOVE = "_NET_WM_STATE_ABOVE"
)

// connectToXServer connects to the X server and returns the connection and screen information.
func connectToXServer() (*xgb.Conn, *xproto.ScreenInfo) {
	x, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("Failed to connect to X server: %v", err)
	}
	setup := xproto.Setup(x)
	screen := setup.DefaultScreen(x)
	return x, screen
}

// internAtom fetches an atom by name from the X server.
func internAtom(x *xgb.Conn, atomName string) xproto.Atom {
	reply, err := xproto.InternAtom(x, true, uint16(len(atomName)), atomName).Reply()
	if err != nil {
		log.Fatalf("Failed to intern atom %s: %v", atomName, err)
	}
	return reply.Atom
}

// runX11Operation runs an operation on an X11 window.
func runX11Operation(win fyne.Window, operation func(x11Context driver.X11WindowContext, x *xgb.Conn, screen *xproto.ScreenInfo)) {
	nativeWin, ok := win.(driver.NativeWindow)
	if !ok {
		panic("this will never happen for a top-level window")
	}

	nativeWin.RunNative(func(ctx any) {
		if runtime.GOOS != "linux" {
			return
		}

		x11Context := ctx.(driver.X11WindowContext)
		x, screen := connectToXServer()
		defer x.Close()

		operation(x11Context, x, screen)
	})
}

func SetWindowAlwaysOnTop(win fyne.Window) {
	runX11Operation(win, func(x11Context driver.X11WindowContext, x *xgb.Conn, screen *xproto.ScreenInfo) {
		x11Window := x11Context.WindowHandle

		netWmStateAtom := internAtom(x, _NET_WM_STATE)
		netWmStateAboveAtom := internAtom(x, _NET_WM_STATE_ABOVE)

		// Prepare client message
		data := []uint32{
			1, // _NET_WM_STATE_ADD
			uint32(netWmStateAboveAtom),
			0,
			0,
			0,
		}
		event := xproto.ClientMessageEvent{
			Format: 32,
			Window: xproto.Window(x11Window),
			Type:   netWmStateAtom,
			Data:   xproto.ClientMessageDataUnionData32New(data),
		}

		// Send the client message
		err := xproto.SendEventChecked(x, false, screen.Root, xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify, string(event.Bytes())).Check()
		if err != nil {
			log.Fatalf("Failed to send _NET_WM_STATE client message: %v", err)
		}
	})
}

func SetWindowOnTopRightCorner(win fyne.Window) {
	runX11Operation(win, func(x11Context driver.X11WindowContext, x *xgb.Conn, screen *xproto.ScreenInfo) {
		x11Window := x11Context.WindowHandle

		// Calculate top-right position
		width := int16(screen.WidthInPixels) - consts.WindowWidth // Screen width minus window width
		height := int16(0)                                        // Top position

		// Move the window to the top-right corner
		err := xproto.ConfigureWindowChecked(x, xproto.Window(x11Window), xproto.ConfigWindowX|xproto.ConfigWindowY, []uint32{uint32(width), uint32(height)}).Check()
		if err != nil {
			log.Fatalf("Failed to move window: %v", err)
		}
	})
}
