package currency

import "fmt"

type Euro struct {
	cents int64
}

func NewEuro(euros float64) Euro {
	return Euro{cents: int64(euros * 100)}
}

func AddEuros(e, e2 Euro) Euro {
	return NewEuro(float64(e.cents+e2.cents) / 100)
}

func (e Euro) Cents() int64 {
	return e.cents
}

func (e Euro) Cmp(e2 Euro) int {
	if e.cents == e2.cents {
		return 0
	} else if e.cents < e2.cents {
		return -1
	} else {
		return 1
	}
}

func (e Euro) String() string {
	return fmt.Sprintf("€%.2f", float64(e.cents)/100)
}
