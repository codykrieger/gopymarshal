package gopymarshal

import (
	"encoding/binary"
	"errors"
	"math"
)

const (
	CODE_NONE      = 'N' //None
	CODE_INT       = 'i' //integer
	CODE_INT2      = 'c' //integer2
	CODE_FLOAT     = 'g' //float
	CODE_STRING    = 's' //string
	CODE_UNICODE   = 'u' //unicode string
	CODE_TSTRING   = 't' //tstring?
	CODE_TUPLE     = '(' //tuple
	CODE_LIST      = '[' //list
	CODE_DICT      = '{' //dict
	CODE_STOP      = '0'
	DICT_INIT_SIZE = 64
)

var (
	ERR_PARSE        = errors.New("invalid data")
	ERR_UNKNOWN_CODE = errors.New("unknown code")
)

type ByteReader interface {
	ReadByte() (byte, error)
	Read(p []byte) (n int, err error)
}

// Unmarshal data serialized by python
func Unmarshal(r ByteReader) (ret interface{}, retErr error) {
	code, err := r.ReadByte()
	if err != nil {
		retErr = err
		return
	}
	return unmarshal(code, r)
}

func unmarshal(code byte, r ByteReader) (ret interface{}, retErr error) {
	switch code {
	case CODE_NONE:
		ret = nil
	case CODE_INT:
		fallthrough
	case CODE_INT2:
		ret, retErr = readInt32(r)
	case CODE_FLOAT:
		ret, retErr = readFloat64(r)
	case CODE_STRING:
		fallthrough
	case CODE_UNICODE:
		fallthrough
	case CODE_TSTRING:
		ret, retErr = readString(r)
	case CODE_TUPLE:
		fallthrough
	case CODE_LIST:
		ret, retErr = readList(r)
	case CODE_DICT:
		ret, retErr = readDict(r)
	default:
		retErr = ERR_UNKNOWN_CODE
	}

	return
}

func readInt32(r ByteReader) (ret int32, retErr error) {
	var tmp int32
	retErr = ERR_PARSE
	if retErr = binary.Read(r, binary.LittleEndian, &tmp); nil == retErr {
		ret = tmp
	}

	return
}

func readFloat64(r ByteReader) (ret float64, retErr error) {
	retErr = ERR_PARSE
	tmp := make([]byte, 8)
	if num, err := r.Read(tmp); nil == err && 8 == num {
		bits := binary.LittleEndian.Uint64(tmp)
		ret = math.Float64frombits(bits)
		retErr = nil
	}

	return
}

func readString(r ByteReader) (ret string, retErr error) {
	var strLen int32
	strLen = 0
	retErr = ERR_PARSE
	if err := binary.Read(r, binary.LittleEndian, &strLen); nil != err {
		retErr = err
		return
	}

	retErr = nil
	buf := make([]byte, strLen)
	r.Read(buf)
	ret = string(buf)
	return
}

func readList(r ByteReader) (ret []interface{}, retErr error) {
	var listSize int32
	if retErr = binary.Read(r, binary.LittleEndian, &listSize); nil != retErr {
		return
	}

	var code byte
	var err error
	var val interface{}
	ret = make([]interface{}, int(listSize))
	for idx := 0; idx < int(listSize); idx++ {
		code, err = r.ReadByte()
		if nil != err {
			break
		}

		val, err = unmarshal(code, r)
		if nil != err {
			retErr = err
			break
		}
		ret = append(ret, val)
	} //end of read loop

	return
}

func readDict(r ByteReader) (ret map[interface{}]interface{}, retErr error) {
	var code byte
	var err error
	var key interface{}
	var val interface{}
	ret = make(map[interface{}]interface{})
	for {
		code, err = r.ReadByte()
		if nil != err {
			break
		}

		if CODE_STOP == code {
			break
		}

		key, err = unmarshal(code, r)
		if nil != err {
			retErr = err
			break
		}

		code, err = r.ReadByte()
		if nil != err {
			break
		}

		val, err = unmarshal(code, r)
		if nil != err {
			retErr = err
			break
		}
		ret[key] = val
	} //end of read loop

	return
}
