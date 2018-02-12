package main

import (
	"errors"
	"unsafe"

	"github.com/gonutz/ide/w32"

	"github.com/gonutz/d3d9"
)

type d3d9Graphics struct {
	window            uintptr
	d3d               *d3d9.Direct3D
	device            *d3d9.Device
	presentParameters d3d9.PRESENT_PARAMETERS
	deviceIsLost      bool
	rectData          []float32
	rectCount         uint
}

func makeErr(context string, err error) error {
	return errors.New(context + ": " + err.Error())
}

func newD3d9Graphics(window uintptr) (*d3d9Graphics, error) {

	d3d, err := d3d9.Create(d3d9.SDK_VERSION)
	if err != nil {
		return nil, makeErr("d3d9.Create", err)
	}

	var createFlags uint32 = d3d9.CREATE_SOFTWARE_VERTEXPROCESSING
	caps, err := d3d.GetDeviceCaps(d3d9.ADAPTER_DEFAULT, d3d9.DEVTYPE_HAL)
	if err == nil && caps.DevCaps&d3d9.DEVCAPS_HWTRANSFORMANDLIGHT != 0 {
		createFlags = d3d9.CREATE_HARDWARE_VERTEXPROCESSING
	}

	const (
		// TODO find the actual maximum screen size and maybe add some margin to
		// it
		maxScreenW = 2048
		maxScreenH = 2048
	)
	pp := d3d9.PRESENT_PARAMETERS{
		Windowed:         1,
		HDeviceWindow:    d3d9.HWND(window),
		SwapEffect:       d3d9.SWAPEFFECT_DISCARD,
		BackBufferWidth:  maxScreenW,
		BackBufferHeight: maxScreenH,
		BackBufferFormat: d3d9.FMT_A8R8G8B8,
		BackBufferCount:  1,
	}
	device, actualPP, err := d3d.CreateDevice(
		d3d9.ADAPTER_DEFAULT,
		d3d9.DEVTYPE_HAL,
		d3d9.HWND(window),
		createFlags,
		pp,
	)
	if err != nil {
		return nil, makeErr("d3d9.CreateDevice", err)
	}
	pp = actualPP

	if err := setRenderState(device); err != nil {
		return nil, err
	}

	g := &d3d9Graphics{
		window:            window,
		d3d:               d3d,
		device:            device,
		presentParameters: pp,
		deviceIsLost:      false,
	}
	return g, nil
}

func (g *d3d9Graphics) close() {
	g.device.Release()
	g.d3d.Release()
}

func (g *d3d9Graphics) rect(x, y, w, h int, argb8 uint32) {
	fx, fy := float32(x), float32(y)
	fx2, fy2 := float32(x+w), float32(y+h)

	var col float32 = *(*float32)(unsafe.Pointer(&argb8))

	// add two triangles for the rectangle
	g.rectData = append(g.rectData,
		fx, fy, 0, 1, col,
		fx2, fy, 0, 1, col,
		fx, fy2, 0, 1, col,

		fx, fy2, 0, 1, col,
		fx2, fy, 0, 1, col,
		fx2, fy2, 0, 1, col)
	g.rectCount++
}

func setRenderState(device *d3d9.Device) error {
	if err := device.SetRenderState(d3d9.RS_CULLMODE, d3d9.CULL_CCW); err != nil {
		return makeErr("SetRenderState(RS_CULLMODE, CULL_CCW)", err)
	}
	if err := device.SetRenderState(d3d9.RS_SRCBLEND, d3d9.BLEND_SRCALPHA); err != nil {
		return makeErr("SetRenderState(RS_SRCBLEND, BLEND_SRCALPHA)", err)
	}
	if err := device.SetRenderState(d3d9.RS_DESTBLEND, d3d9.BLEND_INVSRCALPHA); err != nil {
		return makeErr("SetRenderState(RS_DESTBLEND, BLEND_INVSRCALPHA)", err)
	}
	if err := device.SetRenderState(d3d9.RS_ALPHABLENDENABLE, 1); err != nil {
		return makeErr("SetRenderState(RS_ALPHABLENDENABLE, 1)", err)
	}
	return nil
}

func (g *d3d9Graphics) present() error {
	const (
		rectFmt    = d3d9.FVF_XYZRHW | d3d9.FVF_DIFFUSE
		rectStride = 20
	)

	if g.deviceIsLost {
		_, err := g.device.Reset(g.presentParameters)
		if err != nil {
			// the device is not yet ready for rendering again, this error is
			// not fatal, it may take some frames until the device is ready
			// again
			return nil
		} else {
			if err := setRenderState(g.device); err != nil {
				return err
			}
			g.deviceIsLost = false
		}
	}

	r, ok := w32.GetClientRect(g.window)
	if !ok {
		return errors.New("unable to query window size")
	}
	windowW := r.Right - r.Left
	windowH := r.Bottom - r.Top

	err := g.device.SetViewport(
		d3d9.VIEWPORT{0, 0, uint32(windowW), uint32(windowH), 0, 1},
	)
	if err != nil {
		return err
	}

	if err := g.device.SetFVF(rectFmt); err != nil {
		return err
	}

	// render the scene
	if err := g.device.BeginScene(); err != nil {
		return err
	}

	if err := g.device.DrawPrimitiveUP(
		d3d9.PT_TRIANGLELIST,
		g.rectCount*2, // every rectangle has two triangles
		uintptr(unsafe.Pointer(&g.rectData[0])),
		rectStride,
	); err != nil {
		return err
	}

	g.rectData = g.rectData[0:0]
	g.rectCount = 0

	if err := g.device.EndScene(); err != nil {
		return err
	}

	presentErr := g.device.Present(
		&d3d9.RECT{0, 0, windowW, windowH},
		nil, 0, nil,
	)
	if presentErr != nil {
		if presentErr.Code() == d3d9.ERR_DEVICELOST {
			g.deviceIsLost = true
			return nil
		} else {
			return presentErr
		}
	}

	return nil
}
