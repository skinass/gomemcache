package bin

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/skinass/gomemcache/memcache/types"
)

const ProtoType = "binary"

var DefaultBinCommander = &cmdRunner{}

type cmdRunner struct{}

func (r *cmdRunner) ProtoType() string {
	return ProtoType
}

func (r *cmdRunner) IsAuthSupported() bool {
	return true
}

func (r *cmdRunner) Auth(rw *bufio.ReadWriter, username, password string) error {
	s, err := r.authList(rw)
	if err != nil {
		return err
	}

	switch {
	case strings.Index(s, "PLAIN") != -1:
		return r.authPlain(rw, username, password)
	}

	return fmt.Errorf("memcache: unknown auth types %q", s)
}

func (r *cmdRunner) authPlain(rw *bufio.ReadWriter, username, password string) error {
	m := &msg{
		header: header{
			Op: opAuthStart,
		},

		key: "PLAIN",
		val: []byte(fmt.Sprintf("\x00%s\x00%s", username, password)),
	}

	return sendRecv(rw, m)
}

func (r *cmdRunner) authList(rw *bufio.ReadWriter) (string, error) {
	m := &msg{
		header: header{
			Op: opAuthList,
		},
	}

	err := sendRecv(rw, m)
	return string(m.val), err
}

func (r *cmdRunner) Get(rw *bufio.ReadWriter, keys []string, cb func(*types.Item)) error {
	var err error
	for _, key := range keys {
		if eg := r.getOne(rw, key, cb); eg != nil && eg != types.ErrCacheMiss {
			err = eg
		}
	}

	return err
}

func (r *cmdRunner) getOne(rw *bufio.ReadWriter, key string, cb func(*types.Item)) error {
	var flags uint32
	m := &msg{
		header: header{
			Op:  opGet,
			CAS: uint64(0),
		},
		oextras: []interface{}{&flags},
		key:     key,
	}
	err := sendRecv(rw, m)
	if err != nil {
		return err
	}
	cb(&types.Item{
		Key:   key,
		Value: m.val,
		Casid: m.CAS,
		Flags: flags,
	})
	return nil
}

func (r *cmdRunner) Populate(rw *bufio.ReadWriter, verb types.Verb, item *types.Item) error {
	op := verbToOp(verb)
	var ocas uint64
	if verb == types.Cas {
		ocas = item.Casid
	}

	m := &msg{
		header: header{
			Op:  op,
			CAS: ocas,
		},
		iextras: []interface{}{item.Flags, uint32(item.Expiration)},
		key:     item.Key,
		val:     item.Value,
	}

	err := sendRecv(rw, m)
	if err == types.ErrCASConflict && verb == "add" || verb == "replace" {
		return types.ErrNotStored
	}
	return err
}

func (r *cmdRunner) Delete(rw *bufio.ReadWriter, key string) error {
	return r.DeleteCas(rw, key, 0)
}

func (r *cmdRunner) DeleteCas(rw *bufio.ReadWriter, key string, cas uint64) error {
	m := &msg{
		header: header{
			Op:  opDelete,
			CAS: cas,
		},
		key: key,
	}
	return sendRecv(rw, m)
}

func (r *cmdRunner) DeleteAll(rw *bufio.ReadWriter) error {
	m := &msg{
		header: header{
			Op: opFlush,
		},
	}
	return sendRecv(rw, m)
}

func (r *cmdRunner) FlushAll(rw *bufio.ReadWriter) error {
	return r.DeleteAll(rw)
}

func (r *cmdRunner) Ping(rw *bufio.ReadWriter) error {
	m := &msg{
		header: header{
			Op: opVersion,
		},
	}

	return sendRecv(rw, m)
}

func (r *cmdRunner) Touch(rw *bufio.ReadWriter, keys []string, expiration int32) error {
	exp := uint32(expiration)
	m := &msg{
		header: header{
			Op: opTouch,
		},
		iextras: []interface{}{exp},
	}

	for _, key := range keys {
		m.key = key
		err := sendRecv(rw, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *cmdRunner) IncrDecr(rw *bufio.ReadWriter, verb types.Verb, key string, delta uint64) (uint64, error) {
	op := verbToOp(verb)

	var init uint64
	var exp uint32 = 0xffffffff

	m := &msg{
		header: header{
			Op: op,
		},
		iextras: []interface{}{delta, init, exp},
		key:     key,
	}

	err := sendRecv(rw, m)
	if err != nil {
		return 0, err
	}
	val, err := readInt(string(m.val))
	return val, nil
}

func (r *cmdRunner) LegalKey(key string) bool {
	return true
}

func readInt(b string) (uint64, error) {
	switch len(b) {
	case 8: // 64 bit
		return uint64(uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
			uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56), nil
	}

	return 0, fmt.Errorf("memcache: error parsing int %s", b)
}

func send(rw *bufio.ReadWriter, m *msg) error {
	m.Magic = magicSend
	m.ExtraLen = sizeOfExtras(m.iextras)
	m.KeyLen = uint16(len(m.key))
	m.BodyLen = uint32(m.ExtraLen) + uint32(m.KeyLen) + uint32(len(m.val))
	// m.Opaque = sc.opq
	// sc.opq++

	b := bytes.NewBuffer(nil)
	// Request
	err := binary.Write(b, binary.BigEndian, m.header)
	if err != nil {
		return err
	}

	for _, e := range m.iextras {
		err = binary.Write(b, binary.BigEndian, e)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(b, m.key)
	if err != nil {
		return err
	}

	_, err = b.Write(m.val)
	if err != nil {
		return err
	}

	_, err = rw.Write(b.Bytes())
	if err != nil {
		return err
	}
	return rw.Flush()
}

func recv(r *bufio.Reader, m *msg) error {
	err := binary.Read(r, binary.BigEndian, &m.header)
	if err != nil {
		return err
	}

	bd := make([]byte, m.BodyLen)
	_, err = io.ReadFull(r, bd)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(bd)

	if m.ResvOrStatus == 0 && m.ExtraLen > 0 {
		for _, e := range m.oextras {
			err := binary.Read(buf, binary.BigEndian, e)
			if err != nil {
				return err
			}
		}
	}

	m.key = string(buf.Next(int(m.KeyLen)))
	vlen := int(m.BodyLen) - int(m.ExtraLen) - int(m.KeyLen)
	m.val = buf.Next(int(vlen))
	return newError(m.ResvOrStatus)
}

func sendRecv(rw *bufio.ReadWriter, m *msg) error {
	err := send(rw, m)
	if err != nil {
		return err
	}

	return recv(rw.Reader, m)
}

// sizeOfExtras returns the size of the extras field for the memcache request.
func sizeOfExtras(extras []interface{}) (l uint8) {
	for _, e := range extras {
		switch e.(type) {
		default:
			panic(fmt.Sprintf("mc: unknown extra type (%T)", e))
		case uint8:
			l += 8 / 8
		case uint16:
			l += 16 / 8
		case uint32:
			l += 32 / 8
		case uint64:
			l += 64 / 8
		}
	}
	return
}

func verbToOp(verb types.Verb) opCode {
	switch verb {
	case "incr":
		return opIncrement
	case "decr":
		return opDecrement
	case "cas":
		return opSet
	case "set":
		return opSet
	case "add":
		return opAdd
	case "replace":
		return opReplace
	default:
		return opVersion
	}
}
