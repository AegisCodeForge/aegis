package gitlib

import (
	"errors"
	"io"
)

// pkt-line related.
//     https://git-scm.com/docs/protocol-common

var ErrPktLinePrematureEnd = errors.New("pkt-line ended prematurely.")

func ReadPktLine(r io.Reader) (bool, string, error) {
	c := make([]byte, 4)
	n, err := io.ReadFull(r, c)
	if err != nil { return false, "", err }
	if n < 4 { return false, "", ErrPktLinePrematureEnd }
	ss := (int(charHexToDigit(c[0]))<<12) | (int(charHexToDigit(c[1]))<<8) | (int(charHexToDigit(c[2]))<<4) | (int(charHexToDigit(c[3])))
	if ss == 0 { return true, "", nil }
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

