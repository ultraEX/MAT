// index
package me

import (
	"container/list"
	"fmt"
	"sort"
)

//list.Element:
//Value: int64
///------------------------------------------------------------------
type SortRuleType int64

const (
	SORT_RULE_ASC SortRuleType = 0
	SORT_RULE_DES SortRuleType = 1
)

type SortByAsc []*list.Element

func (I SortByAsc) Len() int {
	return len(I)
}

func (I SortByAsc) Less(i, j int) bool {
	return I[i].Value.(int64) < I[j].Value.(int64)
}

func (I SortByAsc) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

type SortByDes []*list.Element

func (I SortByDes) Len() int {
	return len(I)
}

func (I SortByDes) Less(i, j int) bool {
	return I[i].Value.(int64) > I[j].Value.(int64)
}

func (I SortByDes) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

type IndexInt64 struct {
	indexAsc SortByAsc
	indexDes SortByDes
	rule     SortRuleType
}

func (t *IndexInt64) SortByRule() []*list.Element {
	if t.rule == SORT_RULE_ASC {
		sort.Sort(t.indexAsc)
		return t.indexAsc
	} else {
		sort.Sort(t.indexDes)
		return t.indexDes
	}
}

func (t *IndexInt64) Init(rule SortRuleType) {
	t.rule = rule
}

func (t *IndexInt64) dump() {
	var index []*list.Element
	if t.rule == SORT_RULE_ASC {
		index = t.indexAsc
	} else {
		index = t.indexDes
	}

	fmt.Println("IndexInt64 dump:========================\n")
	count := 0
	for _, elem := range index {
		fmt.Print("[", count, "]: ", elem.Value.(int64), "\n")
	}
}

func NewIndexInt64(rule SortRuleType) *IndexInt64 {
	obj := new(IndexInt64)
	obj.Init(rule)
	return obj
}

// 二分查找
func BinarySearchInt64(m []*list.Element, value int64) (target int) {
	return -1

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		if m[mid].Value.(int64) == value {
			return mid
		}
		if value < m[mid].Value.(int64) {
			if left == right {
				return mid - 1
			} else {
				right = mid - 1
				target = right
			}
		} else if value > m[mid].Value.(int64) {
			if left == right {
				return mid
			} else {
				left = mid + 1
				target = left
			}
		}
	}

	return target
}

func binarySearchUeAsc(m []*list.Element, value int64) (target int, res bool) {

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(int64) > value })

	return target - 1, true
}

func binarySearchUeDes(m []*list.Element, value int64) (target int, res bool) {

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(int64) < value })

	return target - 1, true
}

func binarySearchE(m []*list.Element, value int64) (target int, res bool) {
	if len(m) == 0 {
		return -1, false
	}

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		if m[mid].Value.(int64) == value {
			return mid, true
		}
		if value < m[mid].Value.(int64) {
			right = mid - 1
			target = right
		} else if value > m[mid].Value.(int64) {
			left = mid + 1
			target = left
		}
	}

	return -1, false
}

func (t *IndexInt64) AddToIndex(elem *list.Element) (int, bool) {
	slice := []*list.Element{}
	if t.rule == SORT_RULE_ASC {
		slice = t.indexAsc
	} else {
		slice = t.indexDes
	}

	index := -1
	target, suc := binarySearchUeAsc(slice, elem.Value.(int64))
	if suc {
		s := []*list.Element{}
		if target == -1 {
			s = append(s, elem)
			s = append(s, slice[:]...)
			index = 0
		} else if target == len(slice) {
			if slice[target].Value.(int64) != elem.Value.(int64) {
				s = slice
				s = append(s, elem)
				index = target
			} else {
				///fill repetition
				fmt.Print("Cancel ID=", elem.Value.(int64), ") repetition!\n")
				return -1, false
			}
		} else {
			if slice[target].Value.(int64) != elem.Value.(int64) {
				s = append(s, slice[:target+1]...)
				s = append(s, elem)
				s = append(s, slice[target+1:]...)
				index = target + 1
			} else {
				///fill repetition
				fmt.Print("Cancel ID=", elem.Value.(int64), ") repetition!\n")
				return -1, false
			}
		}

		if t.rule == SORT_RULE_ASC {
			t.indexAsc = s
		} else {
			t.indexDes = s
		}
		return index, true
	}

	return -1, false
}

func (t *IndexInt64) RemoveFromIndex(elem *list.Element) bool {
	slice := []*list.Element{}
	if t.rule == SORT_RULE_ASC {
		slice = t.indexAsc
	} else {
		slice = t.indexDes
	}

	target, suc := binarySearchE(slice, elem.Value.(int64))
	if suc {
		s := []*list.Element{}
		if target == 0 {
			s = slice[1:]
		} else {
			s = append(s, slice[:target]...)
			s = append(s, slice[target+1:]...)
		}

		if t.rule == SORT_RULE_ASC {
			t.indexAsc = s
		} else {
			t.indexDes = s
		}
		return true
	}

	return false
}

func (t *IndexInt64) RemoveFromIndexByIndex(index int) bool {
	slice := []*list.Element{}
	if t.rule == SORT_RULE_ASC {
		slice = t.indexAsc
	} else {
		slice = t.indexDes
	}

	s := []*list.Element{}
	if index == 0 {
		s = slice[1:]
	} else {
		s = append(s, slice[:index]...)
		s = append(s, slice[index+1:]...)
	}

	if t.rule == SORT_RULE_ASC {
		t.indexAsc = s
	} else {
		t.indexDes = s
	}
	return true
}

func (t *IndexInt64) SetIndex(index []*list.Element) {
	if t.rule == SORT_RULE_ASC {
		t.indexAsc = index
	} else {
		t.indexDes = index
	}
}

func (t *IndexInt64) IsExist(elem *list.Element) (int, bool) {
	s := []*list.Element{}
	if t.rule == SORT_RULE_ASC {
		s = t.indexAsc
	} else {
		s = t.indexDes
	}
	target, suc := binarySearchE(s, elem.Value.(int64))
	return target, suc
}
