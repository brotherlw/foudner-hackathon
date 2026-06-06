package guardrails

import (
	"fmt"
	"sync"
	"time"
)

type BudgetTracker struct {
	dailyLimit float64
	mu         sync.Mutex
	spent      float64
	resetDay   string
}

func NewBudgetTracker(dailyLimit float64) *BudgetTracker {
	return &BudgetTracker{
		dailyLimit: dailyLimit,
		resetDay:   today(),
	}
}

func (b *BudgetTracker) Remaining() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	return b.dailyLimit - b.spent
}

func (b *BudgetTracker) CanSpend(amount float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	if b.spent+amount > b.dailyLimit {
		return fmt.Errorf("daily budget exceeded: spent %.2f EUR, limit %.2f EUR, requested %.2f EUR", b.spent, b.dailyLimit, amount)
	}
	return nil
}

func (b *BudgetTracker) RecordSpend(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	b.spent += amount
}

func (b *BudgetTracker) maybeReset() {
	day := today()
	if day != b.resetDay {
		b.spent = 0
		b.resetDay = day
	}
}

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}
