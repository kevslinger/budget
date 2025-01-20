package budget_test

import (
	"testing"

	"github.com/kevslinger/budget"
)

func TestAddEuros(t *testing.T) {
	five := 5.0
	threeFifty := 3.50
	expected := five + threeFifty
	fiveEuros := budget.NewEuro(five)
	threeFiftyEuros := budget.NewEuro(threeFifty)
	actualEuros := budget.AddEuros(fiveEuros, threeFiftyEuros)
	expectedEuros := budget.NewEuro(expected)
	if actualEuros.Cmp(expectedEuros) != 0 {
		t.Errorf("Expected %s but got %s", expectedEuros, actualEuros)
	}
}

func TestCmpEuro(t *testing.T) {
	euros := 59.75
	e1 := budget.NewEuro(euros)
	e2 := budget.NewEuro(euros)
	if e1.Cmp(e2) != 0 {
		t.Errorf("Expected e1 %s and e2 %s to be the same", e1.String(), e2.String())
	}
	if e2.Cmp(e1) != 0 {
		t.Errorf("Expected e2 %s and e1 %s to be the same", e2.String(), e1.String())
	}
}
