package myownsanity

type JoinResult[A any, B any] struct {
	From A
	Join []B
}

func LeftJoin[
	A any, B any,
	E func(a A, b B) bool,
	R struct {
		From A
		Join []B
	},
](inputA []A, inputB []B, equality E) []R {
	accumulator := make([]R, len(inputA))
	for i, a := range inputA {
		items := make([]B, 0)
		for _, b := range inputB {
			if equality(a, b) {
				items = append(items, b)
			}
		}
		accumulator[i] = R{
			From: a,
			Join: items,
		}
	}
	return accumulator
}
