package my_bytes

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"lukechampine.com/uint128"
	"math"
	"strconv"
	"strings"
	"time"
)

func DecodeHead(src []byte) (uint, uint, error) {
	if src == nil || len(src) == 0 {
		return 0, 0, errors.New("nil")
	}
	headLen := uint((src[0]&0x3) + 1)
	if uint(len(src)) < headLen {
		return 0, 0, errors.New("dataLen<headLen")
	}
	n := uint(src[0])
	if headLen > 1 {
		n |= uint(src[1]) << 8
	}
	if headLen > 2 {
		n |= uint(src[2]) << 16
	}
	if headLen > 3 {
		n |= uint(src[3]) << 24
	}
	n >>= 2
	realLength := n
	if n <= 0x3F {
		realLength += 1
	} else if n <= 0x3FFF {
		realLength += 2
	} else if n <= 0x3FFFFF {
		realLength += 3
	} else if n <= 0x3FFFFFFF {
		realLength += 4
	}
	return n, realLength, nil
}

// Decode 去掉变长头，返回纯 payload。
func Decode(src []byte) (data []byte, err error) {
	if src == nil {
		return
	}
	headLen := uint((src[0]&0x3) + 1)
	if uint(len(src)) < headLen {
		err = errors.New("dataLen<headLen")
		return
	}
	data = src[headLen:]
	return
}

// Encoder 将 payload 打包成带变长头的帧。
func Encoder(src []byte) (data []byte, err error) {
	if src == nil {
		return
	}
	l := len(src)
	if l <= 0x3F {
		src = append([]byte{byte(l << 2)}, src...)
	} else if l <= 0x3FFF {
		temp := make([]byte, 2)
		binary.LittleEndian.PutUint16(temp, uint16(l<<2)|0x1)
		src = append(temp, src...)
	} else if l <= 0x3FFFFF {
		h := uint32(l<<2) | 0x2
		temp := make([]byte, 2)
		binary.LittleEndian.PutUint16(temp, uint16(h&0xFFFF))
		src = append([]byte{byte(h >> 16)}, src...)
		src = append(temp, src...)
	} else if l <= 0x3FFFFFFF {
		temp := make([]byte, 4)
		binary.LittleEndian.PutUint32(temp, uint32((l<<2)|0x3))
		src = append(temp, src...)
	} else {
		err = errors.New("overflow")
	}
	return src, nil
}

func EncodeAddr(addr string) (bs []byte) {
	bs = make([]byte, 0)
	tm := uint128.From64(uint64(uint32(time.Now().UnixMicro())))
	split := strings.Split(addr, ":")
	if len(split) != 2 {
		return
	}
	pInt, _ := strconv.ParseUint(split[1], 10, 32)
	for _, s := range strings.Split(split[0], ".") {
		parseInt, _ := strconv.ParseUint(s, 10, 8)
		bs = append(bs, byte(parseInt))
	}
	var y uint32
	_ = binary.Read(bytes.NewBuffer(bs), binary.LittleEndian, &y)
	ip := uint128.From64(uint64(y))
	port := uint128.From64(uint64(pInt))
	a := ip.Add(tm).Lsh(49)
	b := tm.Lsh(17)
	c := tm.And(uint128.From64(0xFFFF)).Add(port)
	d := a.Or(b).Or(c)
	bs = make([]byte, 16)
	d.PutBytes(bs) // PutLittleEndian → PutBytes
	nPadding := 0
	for i := 15; i >= 0; i-- {
		if bs[i] == 0 {
			nPadding++
		} else {
			break
		}
	}
	bs = bs[:(16 - nPadding)]
	return bs
}

func DecodeAddr(data []byte) string {
	bs := make([]byte, 16)
	copy(bs, data)
	number := uint128.FromBytes(bs) // MustU128FromLittleEndian → FromBytes
	tm := number.Rsh(17).And(uint128.From64(math.MaxUint32))
	ip := number.Rsh(49).Sub(tm)
	port := uint16(number.And(uint128.From64(0xFFFFFF)).Sub(tm.And(uint128.From64(0xFFFF))).Lo) // AsUint64() → .Lo
	bs = make([]byte, 16)
	ip.PutBytes(bs)
	bs = bs[:4]
	str := ""
	for i, b := range bs {
		str += strconv.Itoa(int(b))
		if i != 3 {
			str += "."
		}
	}
	return fmt.Sprint(str, ":", port)
}