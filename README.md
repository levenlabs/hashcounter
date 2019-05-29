# hashcounter [![Build Status](https://travis-ci.org/levenlabs/hashcounter.svg?branch=master)](https://travis-ci.org/levenlabs/hashcounter) [![GoDoc](https://godoc.org/github.com/levenlabs/hashcounter?status.svg)](https://godoc.org/github.com/levenlabs/hashcounter)

Provides an efficient way to count the number of occurences of unique integers
provided you can tolerate some collisions. Although not as efficient as
HyperLogLog++, it can also be used to just count the number of distinct values.

The exposed functions are not thread-safe.

Do not use this package if you cannot tolerate any collisions. By default the
key will be determined by calling xxhash.Sum64. You can provide your own hash
function as long as it returns a uint64 (like crc64, murmur3, etc). Since there
is some hashing involved, there might be collisions. You've been warned. If
your byte slices can safely fit in a uint64 without losing any bits then feel
free to send a custom hash function using something like binary.Uvarint.

# Usage

```
c := new(hashcounter.C)
c.Add([]byte(`hello`), 1) // record 1 count of "hello"

cnt, ok := c.Get([]byte(`hello`)) // returns 1 and true
```

# Benchmarks

Tradeoffs were made to prefer lower memory over faster execution speed.

```
BenchmarkAdd-12          2000000              1011 ns/op
BenchmarkGet-12          2000000               955 ns/op
BenchmarkGetKey-12       3000000              1197 ns/op
BenchmarkKey-12         20000000                83.8 ns/op
```
