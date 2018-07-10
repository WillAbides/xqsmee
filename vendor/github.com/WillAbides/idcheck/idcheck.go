// Package idcheck generates ids that can be verified using the final byte as a checksum.
package idcheck

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const length = 16
const hashByte = length - 1

var (
	global = newIDChecker()

	hashTable = func() (vals [256]uint8) {
		for n := range vals {
			vals[n] = uint8(n)
		}
		for _, i := range []int{2, 3, 5, 7, 11} {
			for n := range vals {
				vals[n], vals[n/i] = vals[n/i], vals[n]
			}
		}
		return
	}()
)

type (
	//ID is an id
	ID [length]byte

	//IDChecker issues and validates IDs
	IDChecker interface {
		NewID() (*ID, error)
		ValidID(id *ID) bool
	}

	idChecker struct {
		idReader io.Reader
		salt     []byte
	}

	//Option is an option for creating IDs
	Option func(*idChecker)
)

//SetSalt sets the salt to use in hash calculations
func SetSalt(salt string) { global.SetSalt(salt) }
func (c *idChecker) SetSalt(salt string) {
	c.salt = []byte(salt)
}

//SetIDReader Sets the reader used to generate new IDs
func SetIDReader(idReader io.Reader) { global.SetIDReader(idReader) }
func (c *idChecker) SetIDReader(idReader io.Reader) {
	c.idReader = idReader
}

//Reader use reader to create new IDs instead of the default rand.Reader
func Reader(reader io.Reader) Option {
	return func(opts *idChecker) {
		opts.idReader = reader
	}
}

//Salt sets the salt to use in hash calculations
func Salt(salt string) Option {
	return func(opts *idChecker) {
		opts.salt = []byte(salt)
	}
}

//NewIDChecker returns a new IDChecker
func NewIDChecker(options ...Option) IDChecker {
	return newIDChecker(options...)
}

func newIDChecker(options ...Option) *idChecker {
	opts := &idChecker{
		idReader: rand.Reader,
	}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}

//NewID creates a new ID of default length with a random value
func NewID() (*ID, error) { return global.NewID() }
func (c *idChecker) NewID() (*ID, error) {
	idBytes := make([]byte, length)
	_, err := c.idReader.Read(idBytes)
	if err != nil {
		return nil, err
	}
	id := ID{}
	copy(id[:len(idBytes)-1], idBytes)
	c.setHash(&id)

	return &id, nil
}

//ValidID returns true if the ID has the correct length and checksum
func ValidID(id *ID) bool { return global.ValidID(id) }
func (c *idChecker) ValidID(id *ID) bool {
	return id[hashByte] == c.calculateHash(id)
}

//Base64 encodes an ID as a base64 string
func (id *ID) Base64() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

//FromBase64 creates a new ID from a base64 encoded string
func FromBase64(str string) (*ID, error) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(str)
	if err != nil {
		return nil, err
	}
	if len(b) != length {
		return nil, errors.New("str is not the correct length")
	}
	id := ID{}
	copy(id[:length], b)
	return &id, nil
}

func (c *idChecker) setHash(id *ID) {
	id[hashByte] = c.calculateHash(id)
}

func (c *idChecker) calculateHash(id *ID) (hsh byte) {
	for _, d := range append(c.salt, id[:length-1]...) {
		hsh = hashTable[hsh^d]
	}
	return
}
