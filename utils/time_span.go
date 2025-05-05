package utils

import "time"

// TimeSpan is a struct that represents a time span with a start and end time.
type TimeSpan struct {
	Start time.Time
	End   time.Time
}

func SubtractTimeSpan(base TimeSpan, sub TimeSpan) []TimeSpan {
	var result []TimeSpan

	// sub の 終わりが base の始まりより前か同じ or sub の 始まりが base の終わりより後ろか同じ
	/*
		base: ---s####e----
		sub : s#e----------
		-> 	  ---s####e----

		base: ---s####e----
		sub : ----------s#e
		-> 	  ---s####e----
	*/
	if sub.End.Compare(base.Start) <= 0 || sub.Start.Compare(base.End) >= 0 {
		return []TimeSpan{base}
	}

	// sub の 始まりが base の 始まりより後 かつ base の 終わりより前
	// (先述の early return を考慮すれば base の 終わりより前 は明らか)
	/*
		base: ---s#######e-
		sub : ------s#####e
		-> 	  ---s#e-------

		base: ---s#######e-
		sub : ------s###e--
		-> 	  ---s#e-------
	*/
	if sub.Start.After(base.Start) {
		result = append(result, TimeSpan{
			Start: base.Start,
			End:   sub.Start,
		})
	}

	// sub の 終わりが base の 終わりより前 かつ base の 始まりより後
	// (先述の early return を考慮すれば base の 始まりより後 は明らか)
	/*
		base: ---s#######e-
		sub : -----s#e-----
		-> 	  --------s##e-

		base: ---s#######e-
		sub : -s#####e-----
		-> 	  --------s##e-
	*/
	if sub.End.Before(base.End) {
		result = append(result, TimeSpan{
			Start: sub.End,
			End:   base.End,
		})
	}

	return result
}
