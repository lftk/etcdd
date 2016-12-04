package main

import (
	"encoding/binary"
	"io"
)

func Read(r io.Reader) (b []byte, err error) {
	bn := make([]byte, 4)
	_, err = r.Read(bn)
	if err != nil {
		return
	}
	n := binary.BigEndian.Uint32(bn)
	b = make([]byte, n)
	_, err = r.Read(b)
	return
}

func Write(w io.Writer, b []byte) (err error) {
	bn := make([]byte, 4)
	n := uint32(len(b))
	binary.BigEndian.PutUint32(bn, n)
	_, err = w.Write(bn)
	if err == nil {
		_, err = w.Write(b)
	}
	return
}
