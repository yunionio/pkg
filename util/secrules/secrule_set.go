package secrules

type SecurityRuleSet []SecurityRule

func (srs SecurityRuleSet) Len() int {
	return len(srs)
}

func (srs SecurityRuleSet) Swap(i, j int) {
	srs[i], srs[j] = srs[j], srs[i]
}

func (srs SecurityRuleSet) Less(i, j int) bool {
	if srs[i].Priority > srs[j].Priority {
		return true
	} else if srs[i].Priority == srs[j].Priority {
		return srs[i].String() < srs[j].String()
	}
	return false
}
