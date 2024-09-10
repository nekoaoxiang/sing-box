package rule

import (
	"time"
)

type RuleProvider struct {
	format          string
	lastUpdatedTime time.Time
	ruleCount       uint64
}

func (s *RuleProvider) Format() string {
	return s.format
}

func (s *RuleProvider) RuleCount() uint64 {
	return s.ruleCount
}

func (s *RuleProvider) ListUpdatedTime() time.Time {
	return s.lastUpdatedTime
}
