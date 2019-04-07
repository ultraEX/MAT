// //+build ignore

// A golang Skip List Implementation.
// https://github.com/huandu/skiplist/
//
// Copyright 2011, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/skiplist/blob/master/LICENSE

// Package skiplist provides a skip list implementation in Go.
// About skip list: http://en.wikipedia.org/wiki/Skip_list
//
// Skip list is basically an ordered set.
// Following code creates a skip list with int key and adds some values.
//     list := skiplist.New(skiplist.Int)
//
//     // adds some elements
//     list.Set(20, "Hello")
//     list.Set(10, "World")
//     list.Set(40, true)         // value can be any type
//     list.Set(40, 1000)         // replace last element with new value
//
//     // try to find one
//     e := list.Get(10)           // value is the Element with key 10
//     _ = e.Value.(string)        // it's the "World". remember to do type cast
//     v, ok := list.GetValue(20)  // directly get value. ok is false if not exists
//     v2 := list.MustGetValue(10) // directly get value. panic if key doesn't exist
//     notFound := list.Get(15)    // returns nil if key is not found
//
//     // remove element
//     old := list.Remove(40)     // remove found element and returns its pointer
//                                // returns nil if key is not found
//
//     // re-init list. it will make the list empty.
//     list.Init()
//
// Skip list elements have random number of next pointers. The max number (say
// "max level") is configurable.
//
// The variable skiplist.DefaultMaxLevel is controlling global default.
// Changing it will not affect created lists.
//     skiplist.DefaultMaxLevel = 24  // change default to 24
// Max level of a created list can also be changed even if it's not empty.
//     list.SetMaxLevel(10)
// Remember the side effect when changing this max level value.
// Higher max level usually means higher memory consumption.
// See its wikipedia page for more details.
//
// Most comparable built-in types are pre-defined in skiplist, including
//     byte []byte float32 float64 int int16 int32 int64 int8
//     rune string uint uint16 uint32 uint64 uint8 uintptr
// Pre-defined compare function name is similar to the type name, e.g.
// skiplist.Float32 is for float32 key. A special case is skiplist.Bytes is for []byte.
// These functions order key from small to big (say "ascending order").
// There are also reserved order functions with name like skiplist.IntDesc.
//
// For key types out of the pre-defined list, one can write a custom compare function.
//     type GreaterThanFunc func (lhs, rhs interface{}) bool
// Such compare function returns true if lhs > rhs. Note that, if lhs == rhs, compare
// function (let the name is "foo") must act as following.
//     // if lhs == rhs, following expression must be true
//     foo(lhs, rhs) == false && foo(rhs, lhs) == false
// There is another func type named LessThanFunc. It works similar to GreaterThanFunc,
// except the order is big to small.
// Here is a sample to write a compare func.
//     type Foo struct {
//         value int
//     }
//
//     // it generates a score on a given key.
//     // if key1 > key2, then there must be key1.Score() >= key2.Score().
//     // it's optional. it's worth implementing if call compare func is quite expensive.
//     func (f *Foo) Score() float64 {
//         return float64(f.value)
//     }
//
//     var greater skiplist.GreaterThanFunc = func(lhs, rhs interface{}) {
//         return lhs.(Foo).value > rhs.(Foo).value
//     }
//     list := skiplist.New(greater)
//
//     // descending version is a bit different. mind the func type.
//     var less skiplist.LessThanFunc = func(lhs, rhs interface{}) {
//         return lhs.(Foo).value < rhs.(Foo).value
//     })
//     list := skiplist.New(less)
//
// Skiplist uses global rand source in math/rand by default. This rand source acquires a
// lock when generating random number. Replacing it with a lock-free rand source can provide
// slightly better performance. Use SkipList.SetRandSource to change rand source.
package skiplist

import (
	"bytes"
	"math/rand"
	"sync/atomic"
	"unsafe"
)

const PROPABILITY = 0x3FFF

var (
	DefaultMaxLevel int = 32
	defaultSource       = defaultRandSource{}

	Byte GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(byte) > rhs.(byte)
	}
	ByteAscending               = Byte
	ByteAsc                     = Byte
	ByteDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(byte) < rhs.(byte)
	}
	ByteDesc LessThanFunc = ByteDescending

	Float32 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(float32) > rhs.(float32)
	}
	Float32Ascending               = Float32
	Float32Asc                     = Float32
	Float32Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(float32) < rhs.(float32)
	}
	Float32Desc LessThanFunc = Float32Descending

	Float64 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(float64) > rhs.(float64)
	}
	Float64Ascending               = Float64
	Float64Asc                     = Float64
	Float64Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(float64) < rhs.(float64)
	}
	Float64Desc LessThanFunc = Float64Descending

	Int GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int) > rhs.(int)
	}
	IntAscending               = Int
	IntAsc                     = Int
	IntDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int) < rhs.(int)
	}
	IntDesc LessThanFunc = IntDescending

	Int16 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int16) > rhs.(int16)
	}
	Int16Ascending               = Int16
	Int16Asc                     = Int16
	Int16Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int16) < rhs.(int16)
	}
	Int16Desc LessThanFunc = Int16Descending

	Int32 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int32) > rhs.(int32)
	}
	Int32Ascending               = Int32
	Int32Asc                     = Int32
	Int32Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int32) < rhs.(int32)
	}
	Int32Desc LessThanFunc = Int32Descending

	Int64 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int64) > rhs.(int64)
	}
	Int64Ascending               = Int64
	Int64Asc                     = Int64
	Int64Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int64) < rhs.(int64)
	}
	Int64Desc LessThanFunc = Int64Descending

	Int8 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int8) > rhs.(int8)
	}
	Int8Ascending               = Int8
	Int8Asc                     = Int8
	Int8Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(int8) < rhs.(int8)
	}
	Int8Desc LessThanFunc = Int8Descending

	Rune GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(rune) > rhs.(rune)
	}
	RuneAscending               = Rune
	RuneAsc                     = Rune
	RuneDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(rune) < rhs.(rune)
	}
	RuneDesc LessThanFunc = RuneDescending

	String GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(string) > rhs.(string)
	}
	StringAscending               = String
	StringAsc                     = String
	StringDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(string) < rhs.(string)
	}
	StringDesc LessThanFunc = StringDescending

	Uint GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint) > rhs.(uint)
	}
	UintAscending               = Uint
	UintAsc                     = Uint
	UintDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint) < rhs.(uint)
	}
	UintDesc LessThanFunc = UintDescending

	Uint16 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint16) > rhs.(uint16)
	}
	Uint16Ascending               = Uint16
	Uint16Asc                     = Uint16
	Uint16Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint16) < rhs.(uint16)
	}
	Uint16Desc LessThanFunc = Uint16Descending

	Uint32 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint32) > rhs.(uint32)
	}
	Uint32Ascending               = Uint32
	Uint32Asc                     = Uint32
	Uint32Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint32) < rhs.(uint32)
	}
	Uint32Desc LessThanFunc = Uint32Descending

	Uint64 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint64) > rhs.(uint64)
	}
	Uint64Ascending               = Uint64
	Uint64Asc                     = Uint64
	Uint64Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint64) < rhs.(uint64)
	}
	Uint64Desc LessThanFunc = Uint64Descending

	Uint8 GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint8) > rhs.(uint8)
	}
	Uint8Ascending               = Uint8
	Uint8Asc                     = Uint8
	Uint8Descending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uint8) < rhs.(uint8)
	}
	Uint8Desc LessThanFunc = Uint8Descending

	Uintptr GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uintptr) > rhs.(uintptr)
	}
	UintptrAscending               = Uintptr
	UintptrAsc                     = Uintptr
	UintptrDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return lhs.(uintptr) < rhs.(uintptr)
	}
	UintptrDesc LessThanFunc = UintptrDescending

	// the type []byte.
	Bytes GreaterThanFunc = func(lhs, rhs interface{}) bool {
		return bytes.Compare(lhs.([]byte), rhs.([]byte)) > 0
	}
	BytesAscending = Bytes
	BytesAsc       = Bytes
	// the type []byte. reversed order.
	BytesDescending LessThanFunc = func(lhs, rhs interface{}) bool {
		return bytes.Compare(lhs.([]byte), rhs.([]byte)) < 0
	}
	BytesDesc LessThanFunc = BytesDescending
)

// Return true if lhs greater than rhs
type GreaterThanFunc func(lhs, rhs interface{}) bool

// Return true if lhs less than rhs
type LessThanFunc GreaterThanFunc

type defaultRandSource struct{}

type Comparable interface {
	Descending() bool
	Compare(lhs, rhs interface{}) bool
}

type elementNode struct {
	next []*Element
}

type Element struct {
	elementNode
	key, Value interface{}
	score      float64
}

type SkipList struct {
	elementNode
	level      int
	length     int64
	keyFunc    Comparable
	randSource rand.Source
	reversed   bool

	prevNodesCache []*elementNode // a cache for Set/Remove
}

// It is used by skip list using customized key comparing function.
// For built-in functions, there is no need to care of this interface.
//
// Every skip list element with customized key must have a score value
// to indicate its sequence.
// For any two elements with key "k1" and "k2":
// - If Compare(k1, k2) is true, k1.Score() >= k2.Score() must be true.
// - If Compare(k1, k2) is false and k1 doesn't equal to k2, k1.Score() < k2.Score() must be true.
type Scorable interface {
	Score() float64
}

func (r defaultRandSource) Int63() int64 {
	return rand.Int63()
}

func (r defaultRandSource) Seed(seed int64) {
	rand.Seed(seed)
}

func (f GreaterThanFunc) Descending() bool {
	return false
}

func (f GreaterThanFunc) Compare(lhs, rhs interface{}) bool {
	return f(lhs, rhs)
}

func (f LessThanFunc) Descending() bool {
	return true
}

func (f LessThanFunc) Compare(lhs, rhs interface{}) bool {
	return f(lhs, rhs)
}

// Gets the ajancent next element.
func (element *Element) Next() *Element {
	return element.next[0]
}

// Gets next element at a specific level.
func (element *Element) NextLevel(level int) *Element {
	if level >= len(element.next) || level < 0 {
		panic("invalid argument to NextLevel")
	}

	return element.next[level]
}

// Gets key.
func (element *Element) Key() interface{} {
	return element.key
}

// Creates a new skiplist.
// keyFunc is a func checking the "greater than" logic.
// If k1 equals k2, keyFunc(k1, k2) and keyFunc(k2, k1) must both be false.
// For built-in types, keyFunc can be found in skiplist package.
// For instance, skiplist.Int is for the list with int keys.
// By default, the list with built-in type key is in ascending order.
// The keyFunc named as skiplist.IntDesc is for descending key order list.
func New(keyFunc Comparable) *SkipList {
	if DefaultMaxLevel <= 0 {
		panic("skiplist default level must not be zero or negative")
	}

	return &SkipList{
		elementNode:    elementNode{next: make([]*Element, DefaultMaxLevel)},
		prevNodesCache: make([]*elementNode, DefaultMaxLevel),
		level:          DefaultMaxLevel,
		keyFunc:        keyFunc,
		randSource:     defaultSource,
		reversed:       keyFunc.Descending(),
	}
}

// Resets a skiplist and discards all exists elements.
func (list *SkipList) Init() *SkipList {
	list.next = make([]*Element, list.level)
	atomic.StoreInt64(&list.length, 0)
	return list
}

// Sets a new rand source.
//
// Skiplist uses global rand defined in math/rand by default.
// The default rand acquires a global mutex before generating any number.
// It's not necessary if the skiplist is well protected by caller.
func (list *SkipList) SetRandSource(source rand.Source) {
	list.randSource = source
}

// Gets the first element.
func (list *SkipList) Front() *Element {
	return list.next[0]
}

// Gets the first element.
func (list *SkipList) Iterator() *Element {
	return list.next[0]
}

// Gets list length.
func (list *SkipList) Len() int64 {
	return list.length
}

// Sets a value in the list with key.
// If the key exists, change element value to the new one.
// Returns new element pointer.
func (list *SkipList) Set(key, value interface{}) *Element {
	var element *Element

	score := getScore(key, list.reversed)
	prevs := list.getPrevElementNodes(key, score)

	// found an element with the same key, replace its value
	// if element = prevs[0].next[0]; element != nil && !list.keyFunc.Compare(element.key, key) {
	// 	element.Value = value
	// 	return element
	// }

	element = &Element{
		elementNode: elementNode{
			next: make([]*Element, list.randLevel()),
		},
		key:   key,
		score: score,
		Value: value,
	}

	for i := range element.next {
		element.next[i] = prevs[i].next[i]
		prevs[i].next[i] = element
	}

	atomic.AddInt64(&list.length, 1)
	return element
}

func (list *SkipList) SetCon(key, value interface{}) *Element {
	var element *Element

	score := getScore(key, list.reversed)
	prevs := list.getPrevElementNodesForCompare(key, score, 0)

	// found an element with the same key, replace its value
	// if element = prevs[0].next[0]; element != nil && !list.keyFunc.Compare(element.key, key) {
	// 	element.Value = value
	// 	return element
	// }

	element = &Element{
		elementNode: elementNode{
			next: make([]*Element, list.randLevel()),
		},
		key:   key,
		score: score,
		Value: value,
	}

	for i := range element.next {
		element.next[i] = prevs[i].next[i]
	}

	// for i := range element.next {
	for i := 0; i < len(element.next); {
		// prevs[i].next[i] = element
		// addr := unsafe.Pointer(prevs[i].next[i])
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&prevs[i].next[i])), unsafe.Pointer(element.next[i]), unsafe.Pointer(element)) {
			i++
			continue
		} else {
			prevs = list.getPrevElementNodesForCompare(key, score, i)
		}
	}

	atomic.AddInt64(&list.length, 1)
	return element
}

// Gets an element.
// Returns element pointer if found, nil if not found.
func (list *SkipList) Get(key interface{}) *Element {
	var prev *elementNode = &list.elementNode
	var next *Element
	score := getScore(key, list.reversed)

	for i := list.level - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != nil &&
			(score > next.score || (score == next.score && list.keyFunc.Compare(key, next.key))) {
			prev = &next.elementNode
			next = next.next[i]
		}
	}

	if next != nil && score == next.score && !list.keyFunc.Compare(next.key, key) {
		return next
	}

	return nil
}

// Gets a value. It's a short hand for Get().Value.
// Returns value and its existence status.
func (list *SkipList) GetValue(key interface{}) (interface{}, bool) {
	element := list.Get(key)

	if element == nil {
		return nil, false
	}

	return element.Value, true
}

// Gets a value. It will panic if key doesn't exist.
// Returns value.
func (list *SkipList) MustGetValue(key interface{}) interface{} {
	element := list.Get(key)

	if element == nil {
		panic("cannot find key in skiplist")
	}

	return element.Value
}

// Removes an element.
// Returns removed element pointer if found, nil if not found.
func (list *SkipList) Remove(key interface{}) *Element {
	score := getScore(key, list.reversed)
	prevs := list.getPrevElementNodes(key, score)

	// found the element, remove it
	if element := prevs[0].next[0]; element != nil && !list.keyFunc.Compare(element.key, key) {
		for k, v := range element.next {
			prevs[k].next[k] = v
		}

		atomic.AddInt64(&list.length, -1)
		return element
	}

	return nil
}

func (list *SkipList) RemoveCon(key interface{}) *Element {
	score := getScore(key, list.reversed)
	prevs := list.getPrevElementNodesForCompare(key, score, 0)

	// found the element, remove it
	if element := prevs[0].next[0]; element != nil && !list.keyFunc.Compare(element.key, key) {
		// for k, v := range element.next {
		// 	prevs[k].next[k] = v
		// }
		for i := 0; i < len(element.next); {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&prevs[i].next[i])), unsafe.Pointer(element), unsafe.Pointer(element.next[i])) {
				i++
				continue
			} else {
				// prevs = list.getPrevElementNodesForCompare(key, score, i)
				break
			}
		}

		atomic.AddInt64(&list.length, -1)
		return element
	}

	return nil
}

func (list *SkipList) getPrevElementNodes(key interface{}, score float64) []*elementNode {
	var prev *elementNode = &list.elementNode
	var next *Element

	prevs := list.prevNodesCache

	for i := list.level - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != nil &&
			(score > next.score || (score == next.score && list.keyFunc.Compare(key, next.key))) {
			prev = &next.elementNode
			next = next.next[i]
		}

		prevs[i] = prev
	}

	return prevs
}

func (list *SkipList) getPrevElementNodesForCompare(key interface{}, score float64, uplevel int) []*elementNode {
	var prev *elementNode = &list.elementNode
	var next *Element

	prevs := make([]*elementNode, list.level)

	for i := list.level - 1; i >= uplevel; i-- {
		next = prev.next[i]

		for next != nil &&
			(score > next.score || (score == next.score && list.keyFunc.Compare(key, next.key))) {
			prev = &next.elementNode
			next = next.next[i]
		}

		prevs[i] = prev
	}

	return prevs
}

// Gets current max level value.
func (list *SkipList) MaxLevel() int {
	return list.level
}

// Changes skip list max level.
// If level is not greater than 0, just panic.
func (list *SkipList) SetMaxLevel(level int) (old int) {
	if level <= 0 {
		panic("invalid argument to SetLevel")
	}

	old, list.level = list.level, level

	if old == level {
		return
	}

	if old > level {
		list.next = list.next[:level]
		list.prevNodesCache = list.prevNodesCache[:level]
		return
	}

	next := make([]*Element, level)
	copy(next, list.next)
	list.next = next
	list.prevNodesCache = make([]*elementNode, level)

	return
}

func (list *SkipList) randLevel() int {
	l := 1

	for ((list.randSource.Int63() >> 32) & 0xFFFF) < PROPABILITY {
		l++
	}

	if l > list.level {
		l = list.level
	}

	return l
}

func getScore(key interface{}, reversed bool) (score float64) {
	switch t := key.(type) {
	case []byte:
		var result uint64
		data := []byte(t)
		l := len(data)

		// only use first 8 bytes
		if l > 8 {
			l = 8
		}

		for i := 0; i < l; i++ {
			result |= uint64(data[i]) << uint(8*(7-i))
		}

		score = float64(result)

	case float32:
		score = float64(t)

	case float64:
		score = t

	case int:
		score = float64(t)

	case int16:
		score = float64(t)

	case int32:
		score = float64(t)

	case int64:
		score = float64(t)

	case int8:
		score = float64(t)

	case string:
		var result uint64
		data := string(t)
		l := len(data)

		// only use first 8 bytes
		if l > 8 {
			l = 8
		}

		for i := 0; i < l; i++ {
			result |= uint64(data[i]) << uint(8*(7-i))
		}

		score = float64(result)

	case uint:
		score = float64(t)

	case uint16:
		score = float64(t)

	case uint32:
		score = float64(t)

	case uint64:
		score = float64(t)

	case uint8:
		score = float64(t)

	case uintptr:
		score = float64(t)

	case Scorable:
		score = t.Score()
	}

	if reversed {
		score = -score
	}

	return
}
