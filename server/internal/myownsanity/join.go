package myownsanity

type JoinResult[A any, B any] struct {
	From A
	Join []B
}

func LeftJoin[A any, B any](
	inputA []A,
	inputB []B,
	equality func(a A, b B) bool,
) []JoinResult[A, B] {
	accumulator := make([]JoinResult[A, B], len(inputA))
	for i, a := range inputA {
		items := make([]B, 0)
		for _, b := range inputB {
			if equality(a, b) {
				items = append(items, b)
			}
		}
		accumulator[i] = JoinResult[A, B]{
			From: a,
			Join: items,
		}
	}
	return accumulator
}
