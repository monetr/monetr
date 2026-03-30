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

// GroupByV is the same as [GroupBy] but allows for a value selector for the
// group.
func GroupByV[T any, V any, K comparable](
	items []T,
	keySelector func(element T) K,
	valueSelector func(element T) V,
) []Group[K, V] {
	groupOrder := make([]K, 0, len(items))
	groups := map[K][]V{}
	for _, item := range items {
		key := keySelector(item)
		_, ok := groups[key]
		if !ok {
			groupOrder = append(groupOrder, key)
			groups[key] = []V{valueSelector(item)}
			continue
		}

		groups[key] = append(groups[key], valueSelector(item))
	}

	groupsOrdered := make([]Group[K, V], len(groups))
	for i, key := range groupOrder {
		groupsOrdered[i] = Group[K, V]{
			Key:   key,
			Items: groups[key],
		}
	}
	return groupsOrdered
}

// GroupByMap is the same as [GroupBy] but returns a map instead, this is better
// when the order of the groups themsevles does not matter. The order of the
// items in the group is still preserved relative to the order of the items
// actaully provided to this function though.
func GroupByMap[T any, K comparable](
	items []T,
	keySelector func(element T) K,
) map[K][]T {
	groups := map[K][]T{}
	for _, item := range items {
		key := keySelector(item)
		_, ok := groups[key]
		if !ok {
			groups[key] = []T{item}
			continue
		}

		groups[key] = append(groups[key], item)
	}
	return groups
}

// GroupByMapV is the same as [GroupByMap] except this function allows you to
// specify your own value selector.
func GroupByMapV[T any, V any, K comparable](
	items []T,
	keySelector func(element T) K,
	valueSelector func(element T) V,
) map[K][]V {
	groups := map[K][]V{}
	for _, item := range items {
		key := keySelector(item)
		_, ok := groups[key]
		if !ok {
			groups[key] = []V{valueSelector(item)}
			continue
		}

		groups[key] = append(groups[key], valueSelector(item))
	}
	return groups
}
