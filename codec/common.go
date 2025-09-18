package main

type WireType uint8

const (
	WireTypeVarint WireType = 0
	WireTypeI64    WireType = 1
	WireTypeLen    WireType = 2
	WireTypeSGroup WireType = 3
	WireTypeEGroup WireType = 4
	WireTypeI32    WireType = 5
)
