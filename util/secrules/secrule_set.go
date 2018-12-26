package secrules

type SecurityRuleSet []SecurityRule

func (v SecurityRuleSet) Len() int {
	return len(v)
}

func (v SecurityRuleSet) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v SecurityRuleSet) Less(i, j int) bool {
	if v[i].Priority > v[j].Priority {
		return true
	} else if v[i].Priority == v[j].Priority {
		return v[i].String() < v[j].String()
	}
	return false
}
