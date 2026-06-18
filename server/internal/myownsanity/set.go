package myownsanity

import "encoding/json"

var (
	_ json.Marshaler   = Set[any]{}
	_ json.Unmarshaler = Set[any]{}
)

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](initialValues ...T) Set[T] {
	set := Set[T]{}
	for _, item := range initialValues {
		set.Add(item)
	}
	return set
}

func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) Add(item T) Set[T] {
	s[item] = struct{}{}
	return s
}

func (s Set[T]) Remove(item T) Set[T] {
	delete(s, item)
	return s
}

// MarshalJSON implements [json.Marshaler].
func (s Set[T]) MarshalJSON() ([]byte, error) {
	data := make([]T, 0, len(s))
	for item, _ := range s {
		data = append(data, item)
	}
	return json.Marshal(data)
}

// UnmarshalJSON implements [json.Unmarshaler].
func (s Set[T]) UnmarshalJSON(input []byte) error {
	data := make([]T, 0)
	if err := json.Unmarshal(input, &data); err != nil {
		return err
	}

	for _, item := range data {
		s.Add(item)
	}

	return nil
}
