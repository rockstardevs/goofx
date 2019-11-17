package goofx

import "sync"

var aggregatesMap map[string]struct{}
var initAggegatesMap sync.Once

// GetAggregates returns the singleton aggregates map instance.
func GetAggregates() map[string]struct{} {
	initAggegatesMap.Do(func() {
		var aggregates = []string{
			"OFX",
			"SIGNONMSGSRSV1", "SONRS", "STATUS", "FI",
			"BANKMSGSRSV1", "STMTTRNRS", "STMTRS", "BANKACCTFROM",
			"BANKTRANLIST", "STMTTRN", "LEDGERBAL", "AVAILBAL",
		}
		aggregatesMap = make(map[string]struct{}, len(aggregates))
		for _, a := range aggregates {
			aggregatesMap[a] = struct{}{}
		}
	})
	return aggregatesMap
}

// IsAggregate returns true if the given tag is a know aggregate tag.
func IsAggregate(tag string) bool {
	_, found := GetAggregates()[tag]
	return found
}
