package util

type Set struct {
	memory map[string]interface{}
}

func NewSet() *Set {
	return &Set{
		memory: map[string]interface{}{},
	}
}

func (set *Set) Add(e string) bool {
	if set.Contains(e) {
		return false
	}
	set.memory[e] = nil
	return true
}

func (set *Set) Remove(e string) bool {
	if set.Contains(e) {
		delete(set.memory, e)
		return true
	}
	return false
}

func (set *Set) Clear(e string) {
	set.memory = map[string]interface{}{}
}

func (set *Set) Contains(e string) bool {
	_, ok := set.memory[e]
	return ok
}
