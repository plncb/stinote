package window

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
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

func SetWindowOnTopRightCorner(win fyne.Window, width int) {
	runX11Operation(win, func(x11Context driver.X11WindowContext, x *xgb.Conn, screen *xproto.ScreenInfo) {
		x11Window := x11Context.WindowHandle

		// Initialize RandR extension
		err := randr.Init(x)
		if err != nil {
			log.Fatalf("Failed to initialize RandR: %v", err)
		}

		// Get monitor informations
		resources, err := randr.GetScreenResourcesCurrent(x, screen.Root).Reply()
		if err != nil {
			log.Fatalf("Failed to get RandR screen resources: %v", err)
		}

		// Get focused window
		focusedWindowReply, err := xproto.GetInputFocus(x).Reply()
		if err != nil {
			log.Fatalf("Failed to get focused window: %v", err)
		}
		focusedWindow := focusedWindowReply.Focus

		// Iterate through monitor configurations
		for _, ctrc := range resources.Crtcs {
			crtcInfo, err := randr.GetCrtcInfo(x, ctrc, 0).Reply()
			if err != nil {
				log.Fatalf("Failed to get crtc info: %v", err)
			}

			// Only use active monitor
			if crtcInfo.Width != 0 && crtcInfo.Height != 0 {
				// Only use focused window
				if windowInMonitor(focusedWindow, x, crtcInfo) {
					width := int16(crtcInfo.X + int16(crtcInfo.Width) - int16(width))
					height := int16(crtcInfo.Y)
					log.Printf("width %v", width)

					// Move the window to the top-right corner
					err = xproto.ConfigureWindowChecked(
						x,
						xproto.Window(x11Window),
						xproto.ConfigWindowX|xproto.ConfigWindowY,
						[]uint32{uint32(width), uint32(height)},
					).Check()
					if err != nil {
						log.Fatalf("Failed to move window: %v", err)
					}
				}
			}
		}

	})
}

// Helper function to check if a window is within a monitor's CRTC
func windowInMonitor(window xproto.Window, x *xgb.Conn, crtcInfo *randr.GetCrtcInfoReply) bool {
	windowGeom, err := xproto.GetGeometry(x, xproto.Drawable(window)).Reply()
	if err != nil {
		log.Fatalf("Failed to get window geometry: %v", err)
	}

	// Check if the window is within the bounds of the monitor (crtc)
	return int16(windowGeom.X) >= crtcInfo.X &&
		int16(windowGeom.X) <= crtcInfo.X+int16(crtcInfo.Width) &&
		int16(windowGeom.Y) >= crtcInfo.Y &&
		int16(windowGeom.Y) <= crtcInfo.Y+int16(crtcInfo.Height)
}
