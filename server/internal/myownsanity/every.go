package myownsanity

// Every is a function that takes an array of booleans, if any of the boolean
// values are false then this will return false. If all of them are true then
// this will return true. This can be used to easily evaluate several predicate
// functions and their results.
func Every(tests ...bool) bool {
	for _, test := range tests {
		if !test {
			return false
		}
	}
	return true
}
