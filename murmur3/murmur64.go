package murmur3

import "hash"

// Make sure interfaces are correctly implemented.
var (
	_ hash.Hash   = new(digest64)
	_ hash.Hash64 = new(digest64)
	_ bmixer      = new(digest64)
)

// digest64 is half a digest128.
type digest64 digest128

// New64 returns a 64-bit hasher
func New64() hash.Hash64 { return New64WithSeed(0) }

// New64WithSeed returns a 64-bit hasher set with explicit seed value
func New64WithSeed(seed uint32) hash.Hash64 {
	d := (*digest64)(New128WithSeed(seed).(*digest128))
	return d
}

func (d *digest64) Sum(b []byte) []byte {
	h1 := d.Sum64()
	return append(b,
		byte(h1>>56), byte(h1>>48), byte(h1>>40), byte(h1>>32),
		byte(h1>>24), byte(h1>>16), byte(h1>>8), byte(h1))
}

func (d *digest64) Sum64() uint64 {
	h1, _ := (*digest128)(d).Sum128()
	return h1
}

// Sum64 returns the MurmurHash3 sum of data. It is equivalent to the
// following sequence (without the extra burden and the extra allocation):
//
//	hasher := New64()
//	hasher.Write(data)
//	return hasher.Sum64()
func Sum64(data []byte) uint64 { return Sum64WithSeed(data, 0) }

// Sum64WithSeed returns the MurmurHash3 sum of data. It is equivalent to the
// following sequence (without the extra burden and the extra allocation):
//
//	hasher := New64WithSeed(seed)
//	hasher.Write(data)
//	return hasher.Sum64()
func Sum64WithSeed(data []byte, seed uint32) uint64 {
	d := &digest128{h1: uint64(seed), h2: uint64(seed)}
	d.seed = seed
	d.tail = d.bmix(data)
	d.clen = len(data)
	h1, _ := d.Sum128()
	return h1
}
