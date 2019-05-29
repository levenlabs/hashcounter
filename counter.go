// Package hashcounter provides an efficient way to count the number of
// occurences of unique integers. The exposed functions are not thread-safe.
//
// Do not use this package if you cannot tolerate any collisions. By default
// the key will be determined by calling xxhash.Sum64. You can provide your
// own hash function as long as it returns a uint64 (like crc64, murmur3, etc).
// Since there is some hashing involved, there might be collisions. You've been
// warned. If your byte slices can safely fit in a uint64 without losing any
// bits then feel free to send a custom hash function using something like
// binary.Uvarint.
//
// You can initialize a new C using New or just using new(hashcounter.C).
package hashcounter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cespare/xxhash"
)

const (
	part1Size = 16
	part1Bits = (1 << part1Size) - 1
	idSize    = 64 - part1Size
	idBits    = (1 << idSize) - 1
)

// C holds an array of unique values and an associative count
type C struct {
	arr  [1 << part1Size][]uint64
	hash func([]byte) uint64
}

// New returns a new instance of C
func New() *C {
	return new(C)
}

// NewWithHash returns a new instance of C with the provided hash function
func NewWithHash(fn func([]byte) uint64) *C {
	return &C{
		hash: fn,
	}
}

// Key returns the uint64 key for the given bytes
func (m *C) Key(k []byte) uint64 {
	if m.hash != nil {
		return m.hash(k)
	}
	return xxhash.Sum64(k)
}

func (m *C) loc(k uint64) (uint16, uint64) {
	return uint16(k >> (64 - part1Size)), k & idBits
}

func (m *C) add(p1 uint16, id uint64, v uint16) {
	for i := range m.arr[p1] {
		if id == m.arr[p1][i]&idBits {
			v64 := m.arr[p1][i]>>idSize + uint64(v)
			m.arr[p1][i] = v64<<idSize | id
			return
		}
	}
	m.arr[p1] = append(m.arr[p1], id+uint64(v)<<idSize)
}

// Add adds the value to the given bytes
func (m *C) Add(b []byte, v uint16) {
	p1, id := m.loc(m.Key(b))
	m.add(p1, id, v)
}

// Get returns the value of the given bytes and a boolean if it was found
func (m *C) Get(b []byte) (uint16, bool) {
	return m.GetKey(m.Key(b))
}

// GetKey takes a key rather than bytes but otherwise behaves like Get
func (m *C) GetKey(k uint64) (uint16, bool) {
	p1, id := m.loc(k)
	for i := range m.arr[p1] {
		if id == m.arr[p1][i]&idBits {
			return uint16(m.arr[p1][i] >> idSize), true
		}
	}
	return 0, false
}

// Range calls the given function for every value in the map and continues
// looping until the given bool. The returned key is going to be the result
// of Key(bytes). If you want the key to be reversable, you must pass a hash
// function to NewWithHash that allows you to reverse the operation.
func (m *C) Range(f func(key uint64, value uint16) bool) {
	var key uint64
	for p1 := range m.arr {
		for _, idv := range m.arr[p1] {
			key = uint64(p1)<<(64-part1Size) | idv&idBits
			if !f(key, uint16(idv>>idSize)) {
				return
			}
		}
	}
}

// Len returns a count of all of the keys
func (m *C) Len() int {
	l := 0
	for p1 := range m.arr {
		l += len(m.arr[p1])
	}
	return l
}

// Reset removes all of the keys and returns C to it's empty state
func (m *C) Reset() {
	for p1 := range m.arr {
		m.arr[p1] = nil
	}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (m *C) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write([]byte{1}) // version
	b := make([]byte, binary.MaxVarintLen64)
	for p1 := range m.arr {
		l := len(m.arr[p1])
		if l < 1 {
			continue
		}
		// if part1Size changes then we'll need to change this
		binary.BigEndian.PutUint16(b, uint16(p1))
		buf.Write(b[:2])

		i := binary.PutUvarint(b, uint64(l))
		buf.Write(b[:i])

		for _, idv := range m.arr[p1] {
			binary.BigEndian.PutUint64(b, idv)
			buf.Write(b[:8])
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (m *C) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.New("empty byte slice")
	}
	if b[0] != 1 {
		return fmt.Errorf("unexpected version: %d", b[0])
	}
	b = b[1:]
	for len(b) > 0 {
		p1 := binary.BigEndian.Uint16(b)
		b = b[2:]

		l, res := binary.Uvarint(b)
		if res < 1 {
			return fmt.Errorf("error reading length with Uvarint: %d", res)
		}
		b = b[res:]

		m.arr[p1] = make([]uint64, l)
		for i := range m.arr[p1] {
			m.arr[p1][i] = binary.BigEndian.Uint64(b)
			b = b[8:]
		}
	}
	return nil
}

// Merge adds every key from the sent C to the called on C. This assumes the
// hash functions are the same.
func (m *C) Merge(n *C) {
	for p1 := range n.arr {
		if len(n.arr[p1]) < 1 {
			continue
		}

		// if the array is empty on m then just copy n
		if len(m.arr[p1]) == 0 {
			m.arr[p1] = make([]uint64, len(n.arr[p1]))
			copy(m.arr[p1], n.arr[p1])
			continue
		}

		// otherwise loop over each n value and add it to m
		for _, idv := range n.arr[p1] {
			m.add(uint16(p1), idv&idBits, uint16(idv>>idSize))
		}
	}
}
