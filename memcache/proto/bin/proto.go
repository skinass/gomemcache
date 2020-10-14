package bin

import "github.com/skinass/gomemcache/memcache/types"

// Status Codes that may be returned (usually as part of an Error).
const (
	StatusOK             = uint16(0)
	StatusNotFound       = uint16(1)
	StatusKeyExists      = uint16(2)
	StatusValueTooLarge  = uint16(3)
	StatusInvalidArgs    = uint16(4)
	StatusValueNotStored = uint16(5)
	StatusNonNumeric     = uint16(6)
	StatusAuthRequired   = uint16(0x20)
	StatusAuthContinue   = uint16(0x21)
	StatusUnknownCommand = uint16(0x81)
	StatusOutOfMemory    = uint16(0x82)
	StatusAuthUnknown    = uint16(0xffff)
	StatusNetworkError   = uint16(0xfff1)
	StatusUnknownError   = uint16(0xffff)
)

// newError takes a status from the server and creates a matching Error.
func newError(status uint16) error {
	switch status {
	case StatusOK:
		return nil
	case StatusNotFound:
		return types.ErrCacheMiss
	case StatusKeyExists:
		return types.ErrCASConflict
	case StatusValueTooLarge:
		return types.ErrValueTooLarge
	case StatusInvalidArgs:
		return types.ErrInvalidArgs
	case StatusValueNotStored:
		return types.ErrValueNotStored
	case StatusNonNumeric:
		return types.ErrNonNumeric
	case StatusAuthRequired:
		return types.ErrAuthRequired

	// we only support PLAIN auth, no mechanism that would make use of auth
	// continue, so make it an error for now for completeness.
	case StatusAuthContinue:
		return types.ErrAuthContinue
	case StatusUnknownCommand:
		return types.ErrUnknownCommand
	case StatusOutOfMemory:
		return types.ErrOutOfMemory
	}
	return types.ErrUnknownError
}

type opCode uint8

// ops
const (
	opGet opCode = opCode(iota)
	opSet
	opAdd
	opReplace
	opDelete
	opIncrement
	opDecrement
	opQuit
	opFlush
	opGetQ
	opNoop
	opVersion
	opGetK
	opGetKQ
	opAppend
	opPrepend
	opStat
	opSetQ
	opAddQ
	opReplaceQ
	opDeleteQ
	opIncrementQ
	opDecrementQ
	opQuitQ
	opFlushQ
	opAppendQ
	opPrependQ
	opVerbosity // verbosity - not implemented in memcached (but other servers)
	opTouch
	opGAT
	opGATQ
	opGATK  = opCode(0x23)
	opGATKQ = opCode(0x24)
)

// Auth Ops
const (
	opAuthList opCode = opCode(iota + 0x20)
	opAuthStart
	opAuthStep
)

// Magic Codes
type MagicCode uint8

const (
	magicSend MagicCode = 0x80
	magicRecv MagicCode = 0x81
)

// Memcache header
type header struct {
	Magic        MagicCode
	Op           opCode
	KeyLen       uint16
	ExtraLen     uint8
	DataType     uint8  // not used, memcached expects it to be 0x00.
	ResvOrStatus uint16 // for request this field is reserved / unused, for
	// response it indicates the status
	BodyLen uint32
	Opaque  uint32 // copied back to you in response message (message id)
	CAS     uint64 // version really
}

// Main Memcache message structure
type msg struct {
	header                // [0..23]
	iextras []interface{} // [24..(m-1)] Command specific extras (In)

	// Idea of this is we can pass in pointers to values that should appear in the
	// response extras in this field and the generic send/recieve code can handle.
	oextras []interface{} // [24..(m-1)] Command specifc extras (Out)

	key string // [m..(n-1)] Key (as needed, length in header)
	val []byte // [n..x] Value (as needed, length in header)
}

// Memcache stats
type mcStats map[string]string
