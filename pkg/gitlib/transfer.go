package gitlib

import (
	"errors"
	"io"
)

// pkt-line related.
//     https://git-scm.com/docs/protocol-common

var ErrPktLineEmptyReadIn = errors.New("input empty")
var ErrPktLinePrematureEnd = errors.New("pkt-line ended prematurely.")

func ReadPktLine(r io.Reader) (bool, string, error) {
	c := make([]byte, 4)
	n, err := io.ReadFull(r, c)
	if n == 0 { return false, "", ErrPktLineEmptyReadIn }
	if err != nil { return false, "", err }
	if n < 4 { return false, "", ErrPktLinePrematureEnd }
	ss := (int(charHexToDigit(c[0]))<<12) | (int(charHexToDigit(c[1]))<<8) | (int(charHexToDigit(c[2]))<<4) | (int(charHexToDigit(c[3])))
	switch ss {
	case 0: // flush
		return true, "0000", nil
	case 1: // v2 delimiter packet (see docs/http-clone.org)
		return true, "0001", nil
	case 2: // v2 response end packet (see docs/http-clone.org)
		return true, "0002", nil
	case 3:
		// not specified (but we're writing this just in case git uses
		// it for something new)
		return true, "0003", nil
	}
	data := make([]byte, ss)
	n, err = io.ReadFull(r, data)
	if n != ss { return false, "", ErrPktLinePrematureEnd }
	return false, string(data), nil
}

var ErrTooLongForPktLine = errors.New("Too long to make into pkt-line")

func ToPktLine(s string) (string, error) {
	l := len(s)
	if l > 65531 { return "", ErrTooLongForPktLine }
	ls := intToHex16(l+4)
	return ls + s + "\n", nil
}

