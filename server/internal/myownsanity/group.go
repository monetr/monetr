package myownsanity

type Group[K comparable, T any] struct {
	Key   K
	Items []T
}

// GroupBy takes an array of items and a key selector function. It then returns
// an array of [Group] structs in the order the items appeared in (key wise).
func GroupBy[T any, K comparable](
	items []T,
	keySelector func(element T) K,
) []Group[K, T] {
	groupOrder := make([]K, 0, len(items))
	groups := map[K][]T{}
	for _, item := range items {
		key := keySelector(item)
		_, ok := groups[key]
		if !ok {
			groupOrder = append(groupOrder, key)
			groups[key] = []T{item}
			continue
		}

		groups[key] = append(groups[key], item)
	}

	groupsOrdered := make([]Group[K, T], len(groups))
	for i, key := range groupOrder {
		groupsOrdered[i] = Group[K, T]{
			Key:   key,
			Items: groups[key],
		}
	}
	return groupsOrdered
}
