package store_test

import (
	"fmt"
	"testing"

	store "github.com/hannahhoward/go-storethehash/internal"
	"github.com/stretchr/testify/require"
)

func TestEncodeKeyPosition(t *testing.T) {
	key := []byte("abcdefg")
	offset := 4326
	size := 64
	encoded := store.EncodeKeyPosition(store.KeyPositionPair{key, store.KeyedBlock{Block: store.Block{Offset: store.Position(offset), Size: store.Size(size)}, KeySize: 14}})
	require.Equal(t,
		encoded,
		[]byte{
			0xe6, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0xe, 0x07, 0x61, 0x62, 0x63, 0x64, 0x65,
			0x66, 0x67,
		},
	)
}

func TestRecordListIterator(t *testing.T) {
	// Create records
	var keys []string
	for i := 0; i < 20; i++ {
		keys = append(keys, fmt.Sprintf("key-%02d", i))
	}

	var expected []store.Record
	for i, key := range keys {
		expected = append(expected, store.Record{
			KeyPositionPair: store.KeyPositionPair{
				Key:   []byte(key),
				Block: store.KeyedBlock{Block: store.Block{Offset: store.Position(i), Size: store.Size(i)}, KeySize: store.KeySize(i)},
			},
			Pos: i * 20,
		})
	}

	// Encode them into records list
	var data []byte
	for _, record := range expected {
		encoded := store.EncodeKeyPosition(record.KeyPositionPair)
		data = append(data, encoded...)
	}

	// The record list have the bits that were used to determine the bucket as prefix
	prefixedData := append([]byte{0, 0, 0, 0}, data...)
	// Verify that it can be correctly iterated over those encoded records
	records := store.NewRecordList(prefixedData)
	recordsIter := records.Iter()
	for _, record := range expected {
		require.False(t, recordsIter.Done())
		require.Equal(t, record, recordsIter.Next())
	}
}

func TestRecordListFindKeyPosition(t *testing.T) {
	// Create data
	keys := []string{"a", "ac", "b", "d", "de", "dn", "nky", "xrlfg"}
	// Encode them into records list
	var data []byte
	for i, key := range keys {
		encoded := store.EncodeKeyPosition(store.KeyPositionPair{[]byte(key), store.KeyedBlock{Block: store.Block{Offset: store.Position(i), Size: store.Size(i)}, KeySize: store.KeySize(i)}})
		data = append(data, encoded...)
	}
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedData := append([]byte{0, 0, 0, 0}, data...)
	records := store.NewRecordList(prefixedData)

	// First key
	pos, prevRecord, hasPrev := records.FindKeyPosition([]byte("ABCD"))
	require.Equal(t, pos, 0)
	require.False(t, hasPrev)

	// Between two keys with same prefix, but first one being shorter
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("ab"))
	require.Equal(t, pos, 15)
	require.Equal(t, prevRecord.Key, []byte("a"))

	// Between to keys with both having a different prefix
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("c"))
	require.Equal(t, pos, 46)
	require.Equal(t, prevRecord.Key, []byte("b"))

	// Between two keys with both having a different prefix and the input key having a
	// different length
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("cabefg"))
	require.Equal(t, pos, 46)
	require.Equal(t, prevRecord.Key, []byte("b"))

	// Between two keys with both having a different prefix (with one character in common),
	// all keys having the same length
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("dg"))
	require.Equal(t, pos, 77)
	require.Equal(t, prevRecord.Key, []byte("de"))

	// Between two keys with both having a different prefix, no charachter in in common and
	// different length (shorter than the input key)
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("hello"))
	require.Equal(t, pos, 93)
	require.Equal(t, prevRecord.Key, []byte("dn"))

	// Between two keys with both having a different prefix, no charachter in in common and
	// different length (longer than the input key)
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("pz"))
	require.Equal(t, pos, 110)
	require.Equal(t, prevRecord.Key, []byte("nky"))

	// Last key
	pos, prevRecord, hasPrev = records.FindKeyPosition([]byte("z"))
	require.Equal(t, pos, 129)
	require.Equal(t, prevRecord.Key, []byte("xrlfg"))
}

// Validate that the new key was properly added
func assertAddKey(t *testing.T, records store.RecordList, key []byte) {
	pos, _, _ := records.FindKeyPosition(key)
	newData := records.PutKeys([]store.KeyPositionPair{{key, store.KeyedBlock{Block: store.Block{Offset: store.Position(773), Size: store.Size(48)}, KeySize: 24}}}, pos, pos)
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedNewData := append([]byte{0, 0, 0, 0}, newData...)
	newRecords := store.NewRecordList(prefixedNewData)
	insertedPos, insertedRecord, _ := newRecords.FindKeyPosition(key)
	require.Equal(t,
		insertedPos,
		pos+store.FileOffsetBytes+store.FileSizeBytes+(store.KeySizeBytes*2)+len(key),
	)
	require.Equal(t, insertedRecord.Key, key)
}

func TestRecordListAddKeyWithoutReplacing(t *testing.T) {
	// Create Data
	keys := []string{"a", "ac", "b", "d", "de", "dn", "nky", "xrlfg"}
	// Encode them into records list
	var data []byte
	for i, key := range keys {
		encoded := store.EncodeKeyPosition(store.KeyPositionPair{[]byte(key), store.KeyedBlock{Block: store.Block{Offset: store.Position(i), Size: store.Size(i)}, KeySize: store.KeySize(i)}})
		data = append(data, encoded...)
	}
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedData := append([]byte{0, 0, 0, 0}, data...)
	records := store.NewRecordList(prefixedData)

	// First key
	assertAddKey(t, records, []byte("ABCD"))

	// Between two keys with same prefix, but first one being shorter
	assertAddKey(t, records, []byte("ab"))

	// Between to keys with both having a different prefix
	assertAddKey(t, records, []byte("c"))

	// Between two keys with both having a different prefix and the input key having a
	// different length
	assertAddKey(t, records, []byte("cabefg"))

	// Between two keys with both having a different prefix (with one character in common),
	// all keys having the same length
	assertAddKey(t, records, []byte("dg"))

	// Between two keys with both having a different prefix, no charachter in in common and
	// different length (shorter than the input key)
	assertAddKey(t, records, []byte("hello"))

	// Between two keys with both having a different prefix, no charachter in in common and
	// different length (longer than the input key)
	assertAddKey(t, records, []byte("pz"))

	// Last key
	assertAddKey(t, records, []byte("z"))
}

// Validate that the previous key was properly replaced and the new key was added.
func assertAddKeyAndReplacePrev(t *testing.T, records store.RecordList, key []byte, newPrevKey []byte) {
	pos, prevRecord, hasPrev := records.FindKeyPosition(key)
	require.True(t, hasPrev)

	keys := []store.KeyPositionPair{{newPrevKey, prevRecord.Block}, {key, store.KeyedBlock{Block: store.Block{Offset: store.Position(773), Size: store.Size(48)}, KeySize: 24}}}
	newData := records.PutKeys(keys, prevRecord.Pos, pos)
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedNewData := append([]byte{0, 0, 0, 0}, newData...)
	newRecords := store.NewRecordList(prefixedNewData)

	// Find the newly added prevKey
	insertedPrevKeyPos, insertedPrevRecord, hasPrev := newRecords.FindKeyPosition(newPrevKey)
	require.True(t, hasPrev)
	require.Equal(t, insertedPrevRecord.Pos, prevRecord.Pos)
	require.Equal(t, insertedPrevRecord.Key, newPrevKey)

	// Find the newly added key
	insertedPos, insertedRecord, hasPrev := newRecords.FindKeyPosition(key)
	require.True(t, hasPrev)
	require.Equal(t,
		insertedPos,
		// The prev key is longer, hence use its position instead of the original one
		insertedPrevKeyPos+store.FileOffsetBytes+store.FileSizeBytes+(2*store.KeySizeBytes)+len(key),
	)
	require.Equal(t, insertedRecord.Key, key)
}

// If a new key is added and it fully contains the previous key, them the previous key needs
// to be updated as well. This is what these tests are about.
func TestRecordListAddKeyAndReplacePrev(t *testing.T) {
	// Create Data
	keys := []string{"a", "ac", "b", "d", "de", "dn", "nky", "xrlfg"}
	// Encode them into records list
	var data []byte
	for i, key := range keys {
		encoded := store.EncodeKeyPosition(store.KeyPositionPair{[]byte(key), store.KeyedBlock{Block: store.Block{Offset: store.Position(i), Size: store.Size(i)}, KeySize: store.KeySize(i)}})
		data = append(data, encoded...)
	}
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedData := append([]byte{0, 0, 0, 0}, data...)
	records := store.NewRecordList(prefixedData)

	// Between two keys with same prefix, but first one being shorter
	assertAddKeyAndReplacePrev(t, records, []byte("ab"), []byte("aa"))

	// Between two keys with same prefix, but first one being shorter. Replacing the previous
	// key which is more than one character longer than the existong one.
	assertAddKeyAndReplacePrev(t, records, []byte("ab"), []byte("aaaa"))

	// Between to keys with both having a different prefix
	assertAddKeyAndReplacePrev(t, records, []byte("c"), []byte("bx"))

	// Between two keys with both having a different prefix and the input key having a
	// different length
	assertAddKeyAndReplacePrev(t, records, []byte("cabefg"), []byte("bbccdd"))

	// Between two keys with both having a different prefix (with one character in common),
	// extending the prev key with an additional character to be distinguishable from the new
	// key
	assertAddKeyAndReplacePrev(t, records, []byte("deq"), []byte("dej"))

	// Last key
	assertAddKeyAndReplacePrev(t, records, []byte("xrlfgu"), []byte("xrlfgs"))
}

func TestRecordListGetKey(t *testing.T) {
	// Create Data
	keys := []string{"a", "ac", "b", "de", "dn", "nky", "xrlfg"}
	// Encode them into records list
	var data []byte
	for i, key := range keys {
		encoded := store.EncodeKeyPosition(store.KeyPositionPair{[]byte(key), store.KeyedBlock{Block: store.Block{Offset: store.Position(i), Size: store.Size(i)}, KeySize: store.KeySize(i)}})
		data = append(data, encoded...)
	}
	// The record list have the bits that were used to determine the bucket as prefix
	prefixedData := append([]byte{0, 0, 0, 0}, data...)
	records := store.NewRecordList(prefixedData)

	// First key
	blk, has := records.Get([]byte("a"))
	require.True(t, has)
	require.Equal(t, blk, store.KeyedBlock{Block: store.Block{Offset: store.Position(0), Size: store.Size(0)}, KeySize: 0})

	// Key with same prefix, but it's the second one
	blk, has = records.Get([]byte("ac"))
	require.True(t, has)
	require.Equal(t, blk, store.KeyedBlock{Block: store.Block{Offset: store.Position(1), Size: store.Size(1)}, KeySize: 1})

	// Key with same length as two other keys, sharing a prefix
	blk, has = records.Get([]byte("de"))
	require.True(t, has)
	require.Equal(t, blk, store.KeyedBlock{Block: store.Block{Offset: store.Position(3), Size: store.Size(3)}, KeySize: 3})

	// Key that is sharing a prefix, but is longer
	blk, has = records.Get([]byte("dngho"))
	require.True(t, has)
	require.Equal(t, blk, store.KeyedBlock{Block: store.Block{Offset: store.Position(4), Size: store.Size(4)}, KeySize: 4})

	// Key that is the last one
	blk, has = records.Get([]byte("xrlfg"))
	require.True(t, has)
	require.Equal(t, blk, store.KeyedBlock{Block: store.Block{Offset: store.Position(6), Size: store.Size(6)}, KeySize: 6})

	// Key that is shorter than the inserted ones cannot match
	blk, has = records.Get([]byte("d"))
	require.False(t, has)

	// Key that is before all keys
	blk, has = records.Get([]byte("ABCD"))
	require.False(t, has)

	// Key that is after all keys
	blk, has = records.Get([]byte("zzzzz"))
	require.False(t, has)

	// Key that matches a prefix of some keys, but doesn't match fully
	blk, has = records.Get([]byte("dg"))
	require.False(t, has)
}
