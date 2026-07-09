package field

// AvailableInMonth reports whether t may be purchased when simulating month n.
// month <= 0 disables filtering (all market fields allowed).
func AvailableInMonth(t Type, month int) bool {
	if month <= 0 {
		return true
	}
	info, ok := Catalog[t]
	if !ok {
		return false
	}
	from := info.AvailableFromMonth
	if from < 1 {
		from = 1
	}
	return month >= from
}

// FilterMarket returns market types available in the given month.
func FilterMarket(market []Type, month int) []Type {
	if month <= 0 {
		return append([]Type(nil), market...)
	}
	out := make([]Type, 0, len(market))
	for _, t := range market {
		if AvailableInMonth(t, month) {
			out = append(out, t)
		}
	}
	return out
}
