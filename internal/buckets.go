package store

// BucketIndex is an index to a bucket
type BucketIndex uint32

// Buckets contains pointers to file offsets
//
// The generic specifies how many bits are used to create the buckets. The number of buckets is
// 2 ^ bits.
type Buckets []Position

// NewBuckets returns a list of buckets for the given index size in bits
func NewBuckets(indexSizeBits uint8) (Buckets, error) {
	if indexSizeBits > 32 {
		return nil, ErrIndexTooLarge
	}
	return make(Buckets, 1<<indexSizeBits, 1<<indexSizeBits), nil
}

// Put updates a bucket value
func (b Buckets) Put(index BucketIndex, offset Position) error {
	if int(index) > len(b)-1 {
		return ErrOutOfBounds
	}
	b[int(index)] = offset
	return nil
}

// Get updates returns the value at the given index
func (b Buckets) Get(index BucketIndex) (Position, error) {
	if int(index) > len(b)-1 {
		return 0, ErrOutOfBounds
	}
	return b[int(index)], nil
}

// SizeBuckets contains sizes for all record lists
//
// The generic specifies how many bits are used to create the buckets. The number of buckets is
// 2 ^ bits.
type SizeBuckets []Size

// NewBuckets returns a list of buckets for the given index size in bits
func NewSizeBuckets(indexSizeBits uint8) (SizeBuckets, error) {
	if indexSizeBits > 32 {
		return nil, ErrIndexTooLarge
	}
	return make(SizeBuckets, 1<<indexSizeBits, 1<<indexSizeBits), nil
}

// Put updates a bucket value
func (b SizeBuckets) Put(index BucketIndex, offset Size) error {
	if int(index) > len(b)-1 {
		return ErrOutOfBounds
	}
	b[int(index)] = offset
	return nil
}

// Get updates returns the value at the given index
func (b SizeBuckets) Get(index BucketIndex) (Size, error) {
	if int(index) > len(b)-1 {
		return 0, ErrOutOfBounds
	}
	return b[int(index)], nil
}

// KeySizeBuckets contains sizes for keys in the record list
//
// The generic specifies how many bits are used to create the buckets. The number of buckets is
// 2 ^ bits.
type KeySizeBuckets []KeySize

// NewKeySizeBuckets returns a list of buckets for the given index size in bits
func NewKeySizeBuckets(indexSizeBits uint8) (KeySizeBuckets, error) {
	if indexSizeBits > 32 {
		return nil, ErrIndexTooLarge
	}
	return make(KeySizeBuckets, 1<<indexSizeBits, 1<<indexSizeBits), nil
}

// Put updates a bucket value
func (b KeySizeBuckets) Put(index BucketIndex, keySize KeySize) error {
	if int(index) > len(b)-1 {
		return ErrOutOfBounds
	}
	b[int(index)] = keySize
	return nil
}

// Get updates returns the value at the given index
func (b KeySizeBuckets) Get(index BucketIndex) (KeySize, error) {
	if int(index) > len(b)-1 {
		return 0, ErrOutOfBounds
	}
	return b[int(index)], nil
}
