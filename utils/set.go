package utils

import (
	"bytes"
	"fmt"
)

type Set struct {
	m map[interface{}]bool
}

func NewSet() *Set {
	return &Set{m: make(map[interface{}]bool)}
}

//添加    true 添加成功 false 添加失败
func (set *Set) Add(e interface{}) (b bool) {
	if !set.m[e] {
		set.m[e] = true
		return true
	}
	return false
}

//删除
func (set *Set) Remove(e interface{}) {
	delete(set.m, e)
}

//清除
func (set *Set) Clear() {
	set.m = make(map[interface{}]bool)
}

//是否包含
func (set *Set) Contains(e interface{}) bool {
	return set.m[e]
}

//获取元素数量
func (set *Set) Len() int {
	return len(set.m)
}

//判断两个set时候相同
//true 相同 false 不相同
func (set *Set) Same(other *Set) bool {
	if other == nil {
		return false
	}

	if set.Len() != other.Len() {
		return false
	}

	for k, _ := range set.m {
		if !other.Contains(k) {
			return false
		}
	}

	return true
}

//迭代
func (set *Set) Elements() []interface{} {
	if set == nil {
		fmt.Println("TODO: bv annotation")
	}
	initlen := len(set.m)
	fmt.Println("test1")
	snaphot := make([]interface{}, initlen)

	actuallen := 0

	for k, _ := range set.m {
		if actuallen < initlen {
			snaphot[actuallen] = k
		} else {
			fmt.Println("test")
			snaphot = append(snaphot, k)

		}
		actuallen++
	}

	if actuallen < initlen {
		snaphot = snaphot[:actuallen]
	}

	return snaphot
}

//获取自身字符串
func (set *Set) String() string {
	var buf bytes.Buffer

	buf.WriteString("set{")

	first := true

	for k, _ := range set.m {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}

		buf.WriteString(fmt.Sprintf("%v", k))
	}

	buf.WriteString("}")

	return buf.String()
}
