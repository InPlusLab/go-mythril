package utils

import "sync"

type Set struct {
	Map sync.Map
}

func NewSet() *Set {
	var myMap sync.Map
	return &Set{
		Map: myMap,
	}
}

func (set *Set) Copy() *Set {
	res := NewSet()
	for _, item := range set.Elements() {
		res.Add(item)
	}
	return res
}

func (set *Set) Add(e interface{}) (b bool) {
	_, exist := set.Map.LoadOrStore(e, true)
	return !exist
}

func (set *Set) Remove(e interface{}) {
	set.Map.Delete(e)
}

func (set *Set) Contains(e interface{}) bool {
	_, ok := set.Map.Load(e)
	return ok
}

func (set *Set) Elements() []interface{} {
	res := make([]interface{}, 0)
	set.Map.Range(func(k, v interface{}) bool {
		res = append(res, k)
		return true
	})
	return res
}

func (set *Set) Len() int {
	return len(set.Elements())
}

func (set *Set) Union(other *Set) *Set {
	res := NewSet()
	for _, item := range set.Elements() {
		res.Add(item)
	}
	for _, item := range other.Elements() {
		res.Add(item)
	}
	return res
}

/*
type Set struct {
	//sync.RWMutex
	m map[interface{}]bool
	//m sync.Map
}

func NewSet() *Set {
	//var myMap sync.Map
	return &Set{m: make(map[interface{}]bool)}
	//return &Set{m: myMap}
}

func (set *Set) Copy() *Set {
	res := make(map[interface{}]bool)
	for key, value := range set.m {
		res[key] = value
	}
	return &Set{
		m: res,
	}
}

//添加    true 添加成功 false 添加失败
func (set *Set) Add(e interface{}) (b bool) {
	//set.Lock()
	//defer set.Unlock()
	if !set.m[e] {
		set.m[e] = true
		return true
	}
	return false
	//set.m.Store(e, true)
	//return true
}

//删除
func (set *Set) Remove(e interface{}) {
	//set.Lock()
	//defer set.Unlock()
	delete(set.m, e)
	//set.m.Delete(e)
}

//清除
func (set *Set) Clear() {
	//set.Lock()
	//defer set.Unlock()
	set.m = make(map[interface{}]bool)
}

//是否包含
func (set *Set) Contains(e interface{}) bool {
	//set.RLock()
	//defer set.RUnlock()
	value := set.m[e]
	return value
	//_, ok := set.m.Load(e)
	//return ok
}

//获取元素数量
func (set *Set) Len() int {
	//set.RLock()
	//defer set.RUnlock()
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
	//set.RLock()
	//defer set.RUnlock()

	if set == nil {
		fmt.Println("TODO: bv annotation")
	}
	initlen := len(set.m)

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

func (set *Set) Union(other *Set) *Set {
	res := NewSet()
	for _, item := range set.Elements() {
		res.Add(item)
	}
	for _, item := range other.Elements() {
		res.Add(item)
	}
	return res
}
*/
