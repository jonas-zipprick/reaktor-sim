package seedsearch

// PickUniqueOutcomes returns up to keep ranked outcomes whose board fingerprints
// are not yet in seen. Chosen fingerprints are added to seen.
func PickUniqueOutcomes(
	outcomes []Outcome,
	pick func([]Outcome, int) []Outcome,
	keep int,
	seen map[string]struct{},
) []Outcome {
	if keep < 1 {
		keep = 1
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	ranked := pick(outcomes, len(outcomes))
	chosen := make([]Outcome, 0, keep)
	for _, o := range ranked {
		if _, dup := seen[o.BoardFingerprint]; dup {
			continue
		}
		chosen = append(chosen, o)
		seen[o.BoardFingerprint] = struct{}{}
		if len(chosen) >= keep {
			break
		}
	}
	return chosen
}
