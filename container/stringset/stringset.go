package stringset

type StringSet map[string]bool

func New() StringSet {
	return make(map[string]bool)
}

func (set StringSet) Add(value string) {
	set[value] = true
}

func (set StringSet) Remove(value string) {
	delete(set, value)
}

func (set StringSet) Has(value string) bool {
	return set[value]
}

func (set StringSet) Len() int {
	return len(set)
}
