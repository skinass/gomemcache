package types

type Verb string

const (
	Set     Verb = "set"
	Add          = "add"
	Replace      = "replace"
	Cas          = "cas"
	Incr         = "incr"
	Decr         = "decr"
)
