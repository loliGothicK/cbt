package set

type StringSet map[string]struct{}

func (set StringSet) Add(value string) StringSet {
	set[value] = struct{}{}
	return set
}
func (set StringSet) Delete(value string) StringSet {
	delete(set, value)
	return set
}

func (set StringSet) Range(f func(string) interface{}) []interface{} {
	var ret []interface{}
	for item, _ := range set {
		ret = append(ret, f(item))
	}
	return ret
}
