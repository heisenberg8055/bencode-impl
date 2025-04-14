package bencode

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"sync"
)

var bufioReaderPool sync.Pool

func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func decodeFromReader(r *bufio.Reader) (data interface{}, err error) {
	result, err := unmarshal(r)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshal(data *bufio.Reader) (interface{}, error) {
	ch, err := data.ReadByte()
	if err != nil {
		return nil, err
	}
	switch ch {
	case 'i':
		integerBuffer, err := optimisticReadBytes(data, 'e')
		if err != nil {
			return nil, err
		}
		integerBuffer = integerBuffer[:len(integerBuffer)-1]
		integer, err := strconv.ParseInt(string(integerBuffer), 10, 64)
		if err != nil {
			return nil, err
		}
		return integer, nil
	}
	return "test", nil
}

func optimisticReadBytes(data *bufio.Reader, delim byte) ([]byte, error) {
	buffered := data.Buffered()
	var buffer []byte
	var err error
	if buffer, err = data.Peek(buffered); err != nil {
		return nil, err
	}
	var i int
	var integerbuffer []byte
	if i = bytes.IndexByte(buffer, delim); i >= 0 {
		integerbuffer, err = data.ReadSlice(delim)
	}
	if err != nil {
		return nil, err
	}
	integerBufferLength := len(integerbuffer)
	if integerBufferLength == 1 {
		return nil, fmt.Errorf("integer can't be empty")
	} else if integerBufferLength > 2 && (string(integerbuffer[:2]) == "-0" || string(integerbuffer[:1]) == "0") {
		return nil, fmt.Errorf("zero integer error")
	}
	if err == nil {
		return integerbuffer, nil
	}
	return data.ReadSlice(delim)
}
