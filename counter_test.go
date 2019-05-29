package hashcounter

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	c = new(C)
	m = map[string]uint16{}
)

func init() {
	for i := 0; i < 1e5; i++ {
		// 4 to 16 bytes
		byts := make([]byte, rand.Intn(12)+4)
		rand.Read(byts)
		for j := 0; j < int(byts[0])&0xF; j++ {
			c.Add(byts, 1)
			m[string(byts)] = m[string(byts)] + 1
		}
	}
}

func TestLen(t *testing.T) {
	// make sure length is correct
	assert.Equal(t, len(m), c.Len())
}

func TestGet(t *testing.T) {
	// make sure each key is correct
	for k, v := range m {
		v2, ok := c.Get([]byte(k))
		require.True(t, ok)
		assert.Equal(t, v, v2)
	}
}

func TestRange(t *testing.T) {
	mkeys := map[uint64]uint16{}
	for k, v := range m {
		mkeys[c.Key([]byte(k))] = v
	}
	l := 0
	c.Range(func(k uint64, v uint16) bool {
		assert.Equal(t, mkeys[k], v)
		l++
		return true
	})
	assert.Equal(t, c.Len(), l)
}

func TestMarshalUnmarshal(t *testing.T) {
	b, err := c.MarshalBinary()
	require.NoError(t, err)

	c2 := new(C)
	err = c2.UnmarshalBinary(b)
	require.NoError(t, err)

	require.Equal(t, c.Len(), c2.Len())
	c.Range(func(k uint64, v uint16) bool {
		v2, ok := c2.GetKey(k)
		require.True(t, ok)
		assert.Equal(t, v, v2)
		return true
	})
}

func TestMerge(t *testing.T) {
	c2 := new(C)
	m2 := map[string]uint16{}
	for k, v := range m {
		m2[k] = v
	}

	for i := 0; i < 100; i++ {
		// 4 to 16 bytes
		byts := make([]byte, rand.Intn(12)+4)
		rand.Read(byts)
		for j := 0; j < int(byts[0])&0xF; j++ {
			c2.Add(byts, 1)
			m2[string(byts)] = m2[string(byts)] + 1
		}
	}

	c2.Merge(c)
	require.Equal(t, len(m2), c2.Len())

	// make sure each key is correct
	for k, v := range m2 {
		v2, ok := c2.Get([]byte(k))
		require.True(t, ok)
		assert.Equal(t, v, v2)
	}
}

func TestReset(t *testing.T) {
	c2 := new(C)
	c2.Merge(c)

	c2.Reset()
	assert.Equal(t, 0, c2.Len())
}

func BenchmarkAdd(b *testing.B) {
	vals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		byts := make([]byte, 4)
		rand.Read(byts)
		vals[i] = byts
	}
	c := new(C)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(vals[i], 1)
	}
}

func BenchmarkGet(b *testing.B) {
	c := new(C)
	vals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		byts := make([]byte, 4)
		rand.Read(byts)
		vals[i] = byts
		c.Add(byts, 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(vals[i])
	}
}

func BenchmarkGetKey(b *testing.B) {
	c := new(C)
	vals := make([]uint64, b.N)
	for i := 0; i < b.N; i++ {
		byts := make([]byte, 4)
		rand.Read(byts)
		vals[i] = c.Key(byts)
		c.Add(byts, 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetKey(vals[i])
	}
}

func BenchmarkKey(b *testing.B) {
	c := new(C)
	vals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		byts := make([]byte, 4)
		rand.Read(byts)
		vals[i] = byts
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Key(vals[i])
	}
}
