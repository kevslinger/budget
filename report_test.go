package budget_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevslinger/budget"
)

func TestBudgetReport(t *testing.T) {
	transactions := []budget.Transaction{{Amount: budget.NewEuro(100.0), Description: "Salary"}, {Amount: budget.NewEuro(-50.0), Description: "Groceries"}}
	report := budget.NewBudgetReport("Test", transactions)
	expectedNet := budget.NewEuro(50.0)
	if report.NetIncome != expectedNet {
		t.Errorf("Expected report's Net to be %s, got %s", expectedNet, report.NetIncome)
	}
	expectedTotalIncome := budget.NewEuro(100.0)
	if report.TotalIncome != expectedTotalIncome {
		t.Errorf("Expected report's total income to be %s, got %s", expectedTotalIncome, report.TotalIncome)
	}
	expectedTotalExpense := budget.NewEuro(-50.0)
	if report.TotalExpense != expectedTotalExpense {
		t.Errorf("Expected report's total expense to be %s, got %s", expectedTotalExpense, report.TotalExpense)
	}
}

func TestReadBudgetReportFromFile(t *testing.T) {
	reportName := "Test"
	expected := budget.NewBudgetReport(reportName, []budget.Transaction{{Time: "1", Amount: budget.NewEuro(500), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: budget.NewEuro(-200), Description: "Rent"}})
	actual, err := budget.ReadBudgetReportFromFile(reportName, "testdata/report.csv")
	if err != nil {
		t.Fatalf("Error while reading report from file: %v", err)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected report name %s, got %s", expected.Name, actual.Name)
	}
	if actual.TotalIncome.Cmp(expected.TotalIncome) != 0 {
		t.Errorf("Expected total income %s, got %s", expected.TotalIncome.String(), actual.TotalIncome.String())
	}
	if actual.TotalExpense.Cmp(expected.TotalExpense) != 0 {
		t.Errorf("Expected total expense %s, got %s", expected.TotalExpense.String(), expected.TotalExpense.String())
	}
	if actual.NetIncome.Cmp(expected.NetIncome) != 0 {
		t.Errorf("Expected net income %s, got %s", expected.NetIncome.String(), actual.NetIncome.String())
	}
	if len(actual.Transactions()) != len(expected.Transactions()) {
		t.Errorf("Expected %d transactions, got %d", len(expected.Transactions()), len(actual.Transactions()))
	}
}

func TestCalculateTotalExpensePerDescription(t *testing.T) {
	txs := []budget.Transaction{{Time: "1", Amount: budget.NewEuro(500), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: budget.NewEuro(-200), Description: "Rent"}, {Time: "4", Amount: budget.NewEuro(-50), Description: "Groceries"}, {Time: "5", Amount: budget.NewEuro(2000), Description: "Income"}}
	report := budget.NewBudgetReport("Test", txs)
	expected := map[string]budget.Euro{"Groceries": budget.NewEuro(-75), "Rent": budget.NewEuro(-200)}
	actual := report.CalculateTotalExpensePerDescription()
	if len(actual) != len(expected) {
		t.Fatalf("Expected %d expense categories, got %d", len(expected), len(actual))
	}
	for key, val := range actual {
		if expected[key].Cmp(val) != 0 {
			t.Errorf("Expected %s, got %s", expected[key].String(), val.String())
		}
	}
}

func TestCombineReports(t *testing.T) {
	reportName := "Test"
	report1, err := budget.ReadBudgetReportFromFile(reportName, "testdata/report.csv")
	if err != nil {
		t.Fatalf("Failed while reading report from file: %v", err)
	}
	report2, err := budget.ReadBudgetReportFromFile(reportName, "testdata/report.csv")
	if err != nil {
		t.Fatalf("Failed while reading report from file: %v", err)
	}
	actual := budget.CombineReports(reportName, []budget.Report{report1, report2})

	testTransactions := []budget.Transaction{{Time: "1", Amount: budget.NewEuro(500), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: budget.NewEuro(-200), Description: "Rent"}}
	expected := budget.NewBudgetReport(reportName, append(testTransactions, testTransactions...))
	if actual.Name != expected.Name {
		t.Errorf("Expected report name %s, got %s", expected.Name, actual.Name)
	}
	if actual.TotalIncome.Cmp(expected.TotalIncome) != 0 {
		t.Errorf("Expected total income %s, got %s", expected.TotalIncome.String(), actual.TotalIncome.String())
	}
	if actual.TotalExpense.Cmp(expected.TotalExpense) != 0 {
		t.Errorf("Expected total expense %s, got %s", expected.TotalExpense.String(), expected.TotalExpense.String())
	}
	if actual.NetIncome.Cmp(expected.NetIncome) != 0 {
		t.Errorf("Expected net income %s, got %s", expected.NetIncome.String(), actual.NetIncome.String())
	}
	if len(actual.Transactions()) != len(expected.Transactions()) {
		t.Errorf("Expected %d transactions, got %d", len(expected.Transactions()), len(actual.Transactions()))
	}
}

func TestSortIncomes(t *testing.T) {
	incomes := []budget.Transaction{{Time: "1", Amount: budget.NewEuro(10), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(100), Description: "Income"}, {Time: "3", Amount: budget.NewEuro(50), Description: "Income"}}
	expected := []budget.Transaction{incomes[1], incomes[2], incomes[0]}
	report := budget.NewBudgetReport("Test", incomes)
	actual := report.SortIncomes()
	if len(actual) != len(expected) {
		t.Fatalf("Expected %d incomes, got %d", len(expected), len(actual))
	}
	for idx := range len(actual) {
		if actual[idx].Time != expected[idx].Time || actual[idx].Amount.Cmp(expected[idx].Amount) != 0 || actual[idx].Description != expected[idx].Description {
			t.Errorf("Expected %#v, got %#v", expected[idx], actual[idx])
		}
	}
}

func TestSortExpenses(t *testing.T) {
	firstExpense := budget.Transaction{Time: "January 2025", Amount: budget.NewEuro(-2699.99), Description: "Rent"}
	secondExpense := budget.Transaction{Time: "February 2025", Amount: budget.NewEuro(-2699.99), Description: "Rent"}
	thirdExpense := budget.Transaction{Time: "March 2025", Amount: budget.NewEuro(-2699.99), Description: "Rent"}
	expectedSortedExpenses := []budget.Transaction{firstExpense, secondExpense, thirdExpense}
	report := budget.NewBudgetReport("Test", []budget.Transaction{firstExpense, secondExpense, thirdExpense})
	actualSortedExpenses := report.SortExpenses()
	if len(expectedSortedExpenses) != len(actualSortedExpenses) {
		t.Errorf("Expected %#v got %#v", expectedSortedExpenses, actualSortedExpenses)
	}
	for idx := range len(expectedSortedExpenses) {
		expected, actual := expectedSortedExpenses[idx], actualSortedExpenses[idx]
		if expected.Time != actual.Time || expected.Amount.Cmp(actual.Amount) != 0 || expected.Description != actual.Description {
			t.Errorf("Expected %#v got %#v", expected, actual)
		}
	}
}

func TestTransactions(t *testing.T) {
	expected := []budget.Transaction{{Time: "1", Amount: budget.NewEuro(500), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: budget.NewEuro(-200), Description: "Rent"}}
	report := budget.NewBudgetReport("Test", expected)
	actual := report.Transactions()
	if len(actual) != len(expected) {
		t.Fatalf("Expected %d transactions, got %d", len(expected), len(actual))
	}
	for idx := range len(actual) {
		if actual[idx].Time != expected[idx].Time || actual[idx].Amount.Cmp(expected[idx].Amount) != 0 || actual[idx].Description != expected[idx].Description {
			t.Errorf("Expected %#v, got %#v", expected, actual)
		}
	}
}

func TestSave(t *testing.T) {
	transactions := []budget.Transaction{{Time: "1", Amount: budget.NewEuro(500), Description: "Income"}, {Time: "2", Amount: budget.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: budget.NewEuro(-200), Description: "Rent"}}
	report := budget.NewBudgetReport("Test", transactions)
	expected := `Time,Amount,Description
1,500.00,Income
3,-200.00,Rent
2,-25.00,Groceries`
	buffer := new(bytes.Buffer)
	err := report.WriteCSV(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if buffer.String() != expected {
		t.Errorf("Expected %s, got %s, %s", expected, buffer.String(), cmp.Diff(expected, buffer.String()))
	}
}
