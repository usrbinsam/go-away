package iter

type Iterator[T any] interface {
	Next() (*T, bool)
}

type Iter[T any] struct {
	index int
	items []T
}

func (it *Iter[T]) Next() (*T, bool) {
	if it.index >= len(it.items) {
		return nil, false
	}
	it.index++
	return &it.items[it.index-1], true
}
