package poset

type Item interface {
	String() string
}

type Poset struct {
	index map[string]interface{}
	order StringPoset
}

func NewPoset() *Poset {
	return &Poset{
		index: make(map[string]interface{}),
		order: NewPoset(),
	}
}

func (poset *Poset) Add(pre, post Item) {
	poset.index[post.String()] = post
	poset.index[pre.String()] = pre
	poset.order.Add(pre.String(), post.String())
}

func (poset *Poset) Pop() interface{} {
	key := poset.order.Pop()
	if key == "" {
		return nil
	}
	result := poset.index[key]
	delete(poset.index, key)
	return result
}
