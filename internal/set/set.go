package set

type Set[T comparable] map[T]struct{}

func New[T comparable](values ...T) Set[T] {
	out := Set[T]{}
	for _, value := range values {
		out.Add(value)
	}
	return out
}

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Delete(value T) {
	delete(s, value)
}

func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

func (s Set[T]) Equal(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for value := range s {
		if !other.Has(value) {
			return false
		}
	}
	return true
}
