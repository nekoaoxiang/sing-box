package rule

type RuleProvider struct {
	format    string
	ruleCount uint64
}

func (s *RuleProvider) Format() string {
	return s.format
}

func (s *RuleProvider) RuleCount() uint64 {
	return s.ruleCount
}
