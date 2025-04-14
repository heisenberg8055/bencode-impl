package bencode

import (
	"bufio"
	"bytes"
	"errors"
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
	case 'l':
		list := []interface{}{}
		for {
			c, err2 := data.ReadByte()
			if err2 == nil {
				if c == 'e' {
					return list, nil
				} else {
					data.UnreadByte()
				}
			}
			value, err := unmarshal(data)
			if err != nil {
				return nil, err
			}
			list = append(list, value)
		}
	case 'd':
		dictionary := map[string]interface{}{}
		comp := ""
		for {
			c, err2 := data.ReadByte()
			if err2 == nil {
				if c == 'e' {
					return dictionary, nil
				} else {
					data.UnreadByte()
				}
			}
			value, err := unmarshal(data)
			if err != nil {
				return nil, err
			}
			key, ok := value.(string)
			if !ok {
				return nil, errors.New("bencode: non-string dictionary key")
			}
			if key <= comp {
				return nil, errors.New("bencode: keys not sorted")
			}
			comp = key
			value, err = unmarshal(data)
			if err != nil {
				return nil, err
			}
			_, ok = dictionary[key]
			if ok {
				return nil, errors.New("bencode: duplicate dictionary key")
			}
			dictionary[key] = value
		}
	default:
		data.UnreadByte()
		stringLengthBuffer, err := optimisticReadBytes(data, ':')
		if err != nil {
			return nil, err
		}
		stringLengthBuffer = stringLengthBuffer[:len(stringLengthBuffer)-1]
		stringLength, err := strconv.ParseInt(string(stringLengthBuffer), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse string length")
		}
		buf := make([]byte, stringLength)
		_, err = readAtLeast(data, buf, int(stringLength))
		return string(buf), err
	}
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
	return integerbuffer, nil
}

func readAtLeast(data *bufio.Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, io.ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = data.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
