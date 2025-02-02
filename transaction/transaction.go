package transaction

import "github.com/kevslinger/budget/currency"

// BasicTransaction contains the information to describe a single income or expense
type BasicTransaction struct {
	Amount      currency.Euro
	Description string
	Time        string
}

// PayerTransaction contains the information to describe a single income or expense, including who earned/paid
type PayerTransaction struct {
	Amount      currency.Euro
	Description string
	Time        string
	PaidBy      string
}
