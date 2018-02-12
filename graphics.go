package main

type graphics interface {
	rect(x, y, w, h int, argb8 uint32)
	present() error
}
