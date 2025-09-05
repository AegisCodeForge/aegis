package gitlib

import "bytes"

// this is here because UnreadByte somehow can't unread a whole line.
type trueLineReader struct {
	reader *bytes.Buffer
	buffer []string
}

func newTrueLineReader(r *bytes.Buffer) *trueLineReader {
	return &trueLineReader{
		reader: r,
		buffer: make([]string, 0),
	}
}
func (r *trueLineReader) readLine() (string, error) {
	if len(r.buffer) > 0 {
		rb, e := r.buffer[:len(r.buffer)-1], r.buffer[len(r.buffer)-1]
		r.buffer = rb
		return e, nil
	}
	return r.reader.ReadString('\n')
}
func (r *trueLineReader) unreadLine(l string) {
	r.buffer = append(r.buffer, l)
}
