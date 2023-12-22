package qfx

import "io"

type Tokenizer struct {
	input     io.Reader
	register  []byte
	peakIndex int
	index     int
	line      int
	offset    int
}

func NewQFXTokenizer(input io.Reader) *Tokenizer {
	return &Tokenizer{
		input:     input,
		register:  make([]byte, 1),
		peakIndex: 0,
		index:     0,
		line:      1,
		offset:    0,
	}
}

func (t *Tokenizer) peak() (byte, bool) {
	if t.peakIndex > t.index {
		return t.register[0], true
	}

	t.input.Read(t.register)
}
