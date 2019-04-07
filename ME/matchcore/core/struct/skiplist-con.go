package skiplist

import (
	"bytes"
	"errors"
	"math/rand"
	"sync/atomic"
	"unsafe"
)

var (
	ErrNilKey         = errors.New("nil key")
	ErrNilValue       = errors.New("nil value")
	ErrKeyExists      = errors.New("key already exists")
	ErrNilAction      = errors.New("nil action")
	ErrUnknownFromKey = errors.New("unknown from key")
	ErrUnknownToKey   = errors.New("unknown to key")
	ErrRemoveNilNode  = errors.New("can't remove nil node")
)

type node struct {
	key    []byte
	value  unsafe.Pointer
	next   *node
	marked int32
}

//create new node
func newNode(key []byte, value unsafe.Pointer, next *node, prev *node) *node {
	return &node{
		key:   key,
		value: value,
		next:  next,
	}
}

//delete marked node
func (n *node) deleteMarkedNode(prev, succ *node) {
	if n == prev.next && succ == n.next {
		prev.casNext(n, succ)
	}
}

//cas next
func (n *node) casNext(cmp, succ *node) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next)), unsafe.Pointer(cmp), unsafe.Pointer(succ))
}

type Iterator struct {
	sl       *Skiplist
	lo       []byte
	hi       []byte
	next     *node
	lastNode *node
}

//create iterator
func NewIterator(sl *Skiplist, fromKey, toKey []byte) (*Iterator, error) {
	i := &Iterator{
		sl: sl,
		lo: fromKey,
		hi: toKey,
	}

	if fromKey == nil {
		i.next = sl.findFirstNode()
		return i, nil
	}

	var exactMatch bool
	if i.next, exactMatch = sl.findPrecursorOrNode(fromKey); !exactMatch {
		return nil, ErrUnknownFromKey
	}

	return i, nil
}

//next
func (i *Iterator) Next() bool {
	return i.next != nil
}

//get next node
func (i *Iterator) NextNode() ([]byte, unsafe.Pointer) {
	i.lastNode = i.next

	p := i.next.next
	for p != nil {
		if atomic.LoadInt32(&p.marked) == 0 {
			i.next = p
			break
		}

		p = p.next
	}

	if i.next == i.lastNode || (i.hi != nil && bytes.Compare(i.next.key, i.hi) == 1) {
		i.next = nil
	}

	return i.lastNode.key, i.lastNode.value
}

//remove
func (i *Iterator) Remove() error {
	if i.next == nil {
		return ErrRemoveNilNode
	}

	i.sl.remove(i.lastNode.key, i.lastNode.value)
	return nil
}

type index struct {
	node  *node
	right *index
	down  *index
	level int
}

//create header
func newHeader(key []byte, value unsafe.Pointer, next *node, down, right *index, level int) *index {
	return &index{
		node: &node{
			key:   key,
			value: value,
			next:  next,
		},
		down:  down,
		right: right,
		level: level,
	}
}

//create index
func newIndex(n *node, down, right *index) *index {
	return &index{
		node:  n,
		down:  down,
		right: right,
	}
}

//delete marked node
func (i *index) deleteMarkedNode(succ *index) bool {
	return atomic.LoadInt32(&i.node.marked) == 0 && i.casRight(succ, succ.right)
}

//cas right index
func (i *index) casRight(cmp, right *index) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&i.right)), unsafe.Pointer(cmp), unsafe.Pointer(right))
}

//add index
func (i *index) addIndex(succ, newSucc *index) bool {
	newSucc.right = succ

	return atomic.LoadInt32(&i.node.marked) == 0 && i.casRight(succ, newSucc)
}

type Skiplist struct {
	MaxLevel int
	header   *index
	count    int64
}

//create new skiplist
func NewSkiplist(maxLevel int) *Skiplist {
	return &Skiplist{
		MaxLevel: maxLevel,
		header:   newHeader(nil, nil, nil, nil, nil, 1),
	}
}

//random level
func (sl *Skiplist) randomLevel() int {
	level := 1

	for level < sl.MaxLevel && rand.Intn(4) == 0 {
		level++
	}

	return level
}

//find precursor
func (sl *Skiplist) findPrecursorOrNode(key []byte) (*node, bool) {
	for {
		//from sl.level to level 1
		q := sl.header
		r := q.right

		for {
			if r != nil {
				n := r.node

				//delete marked node
				if atomic.LoadInt32(&n.marked) == 1 {
					if !q.deleteMarkedNode(r) {
						continue
					}
				}

				//compare key
				c := bytes.Compare(n.key, key)
				if c == 0 && atomic.LoadInt32(&n.marked) == 0 {
					return n, true
				}

				//n.key < key, go right
				if c == -1 {
					q = r
					r = q.right
					continue
				}
			}

			//r is nil || n.key > key, to the next level
			d := q.down

			//q is level 0, return the node of q
			if d == nil {
				return q.node, bytes.Equal(q.node.key, key)
			}

			q = d
			r = q.right
		}
	}
}

func (sl *Skiplist) Put(key []byte, value unsafe.Pointer) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if value == nil {
		return nil, ErrNilValue
	}

	return sl.put(key, value, nil, false)
}

func (sl *Skiplist) PutOnlyIfAbsent(key []byte, value unsafe.Pointer) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if value == nil {
		return nil, ErrNilValue
	}

	return sl.put(key, value, nil, true)
}

func (sl *Skiplist) Update(key []byte, action func(unsafe.Pointer) unsafe.Pointer) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if action == nil {
		return nil, ErrNilAction
	}

	return sl.put(key, nil, action, false)
}

//put action
func (sl *Skiplist) put(key []byte, value unsafe.Pointer, action func(unsafe.Pointer) unsafe.Pointer, onlyIfAbsent bool) (unsafe.Pointer, error) {
	for {
		prev, exactMatch := sl.findPrecursorOrNode(key)
		//key exists
		if exactMatch {
			oldValue := prev.value

			if !onlyIfAbsent {
				//value is not nil
				if value != nil {
					if atomic.CompareAndSwapPointer(&(prev.value), oldValue, value) {
						return oldValue, nil
					}

					continue
				}

				//action is not nil
				if action != nil {
					newValue := action(oldValue)
					if atomic.CompareAndSwapPointer(&(prev.value), oldValue, newValue) {
						return oldValue, nil
					}

					continue
				}
			}

			return nil, ErrKeyExists
		}

		//key not exists
		if value == nil {
			return nil, ErrNilValue
		}

		n := prev.next
		for {
			if n != nil {
				succ := n.next

				//check node whether changed
				if n != prev.next {
					break
				}

				//delete marked node
				if atomic.LoadInt32(&n.marked) == 1 {
					n.deleteMarkedNode(prev, succ)
					break
				}

				//compare key
				c := bytes.Compare(n.key, key)

				//if n.key < key, go next
				if c == -1 {
					prev = n
					n = succ
					continue
				}
			}

			//create new node
			nn := newNode(key, value, n, prev)
			if prev.casNext(n, nn) {
				atomic.AddInt64(&sl.count, 1)
			} else {
				break
			}

			//get level
			level := sl.randomLevel()

			//insert index to each level
			sl.insertIndex(nn, level)

			return nil, nil
		}
	}
}

//Get
func (sl *Skiplist) Get(key []byte) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	return sl.get(key), nil
}

//Contains
func (sl *Skiplist) Contains(key []byte) (bool, error) {
	if key == nil {
		return false, ErrNilKey
	}

	return sl.get(key) != nil, nil
}

//get
func (sl *Skiplist) get(key []byte) unsafe.Pointer {
	if n, excatMatch := sl.findPrecursorOrNode(key); excatMatch {
		return n.value
	}

	return nil
}

//reomve
func (sl *Skiplist) Remove(key []byte) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	return sl.remove(key, nil), nil
}

//compareAndRemove
func (sl *Skiplist) CompareAndRemove(key []byte, value unsafe.Pointer) (unsafe.Pointer, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if value == nil {
		return nil, ErrNilValue
	}

	return sl.remove(key, value), nil
}

func (sl *Skiplist) remove(key []byte, value unsafe.Pointer) unsafe.Pointer {
	for {
		n, exactMatch := sl.findPrecursorOrNode(key)

		//key exists
		if exactMatch {
			if (value != nil && value == n.value) || value == nil {
				for atomic.LoadInt32(&n.marked) == 0 {
					atomic.StoreInt32(&n.marked, 1)
				}

				atomic.AddInt64(&sl.count, -1)
				return n.value
			}
		}

		return nil
	}
}

//Size
func (sl *Skiplist) Count() int64 {
	return atomic.LoadInt64(&sl.count)
}

//insert index
func (sl *Skiplist) insertIndex(n *node, level int) {
	h := sl.header
	max := h.level

	var idx *index
	if level <= max {
		for i := 1; i <= level; i++ {
			idx = newIndex(n, idx, nil)
		}

		//add index to skiplist
		sl.addIndex(idx, h, level, max, 1)
		return
	}

	level = max + 1
	idxs := make([]*index, level+1)

	for i := 1; i <= level; i++ {
		idx = newIndex(n, idx, nil)
		idxs[i] = idx
	}

	var k int
	var oldHeader *index
	for {
		oldHeader = sl.header
		oldLevel := oldHeader.level
		if level <= oldLevel {
			k = level
			break
		}

		nh := oldHeader
		oldNode := oldHeader.node
		for j := oldLevel + 1; j <= level; j++ {
			nh = newHeader(oldNode.key, oldNode.value, oldNode.next, nh, idxs[j], j)
		}

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&sl.header)), unsafe.Pointer(oldHeader), unsafe.Pointer(nh)) {
			k = oldLevel
			break
		}
	}

	sl.addIndex(idxs[k], oldHeader, k, max, 2)
}

//add index
func (sl *Skiplist) addIndex(idx, h *index, level, max, pos int) {
	l1 := level

	for {
		l := h.level
		q := h
		r := q.right
		t := idx

		for {
			if r != nil {
				n := r.node
				c := bytes.Compare(idx.node.key, n.key)

				//delete marked node
				if atomic.LoadInt32(&n.marked) == 1 {
					if !q.deleteMarkedNode(r) {
						break
					}
				}

				//find prev
				if c == 1 {
					q = r
					r = r.right
					continue
				}
			}

			if l == l1 {
				//idx is removed
				if atomic.LoadInt32(&idx.node.marked) == 1 {
					//delete marked idx node
					sl.findPrecursorOrNode(idx.node.key)
					return
				}

				if !q.addIndex(r, t) {
					break
				}

				l1--
				if l1 == 0 {
					if atomic.LoadInt32(&idx.node.marked) == 1 {
						sl.findPrecursorOrNode(idx.node.key)
					}

					return
				}
			}

			l--
			if l >= l1 && l < level {
				t = t.down
			}

			q = q.down
			r = q.right
		}
	}
}

//find last level header
func (sl *Skiplist) findBottomHeader() *index {
	h := sl.header
	d := h.down

	for {
		if d == nil {
			return h
		}
		if d.down == nil {
			return d
		}

		d = d.down
	}
}

//find first node
func (sl *Skiplist) findFirstNode() *node {
	h := sl.findBottomHeader()
	r := h.right
	if r == nil {
		return nil
	}

	return r.node
}
