package report_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevslinger/budget/currency"
	"github.com/kevslinger/budget/report"
	"github.com/kevslinger/budget/transaction"
)

func TestBudgetReport(t *testing.T) {
	transactions := []transaction.BasicTransaction{{Amount: currency.NewEuro(100.0), Description: "Salary"}, {Amount: currency.NewEuro(-50.0), Description: "Groceries"}}
	report := report.NewBasicBudgetReport("Test", transactions)
	expectedNet := currency.NewEuro(50.0)
	if report.NetIncome != expectedNet {
		t.Errorf("Expected report's Net to be %s, got %s", expectedNet, report.NetIncome)
	}
	expectedTotalIncome := currency.NewEuro(100.0)
	if report.TotalIncome != expectedTotalIncome {
		t.Errorf("Expected report's total income to be %s, got %s", expectedTotalIncome, report.TotalIncome)
	}
	expectedTotalExpense := currency.NewEuro(-50.0)
	if report.TotalExpense != expectedTotalExpense {
		t.Errorf("Expected report's total expense to be %s, got %s", expectedTotalExpense, report.TotalExpense)
	}
}

func TestReadBudgetReportFromFile(t *testing.T) {
	reportName := "Test"
	expected := report.NewBasicBudgetReport(reportName, []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(500.0), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(-25.0), Description: "Groceries"}, {Time: "3", Amount: currency.NewEuro(-200.0), Description: "Rent"}})
	actual, err := report.ReadDefaultBudgetReportFromFile(reportName, "../testdata/defaultreport.csv")
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
	txs := []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(500.0), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(-25.0), Description: "Groceries"}, {Time: "3", Amount: currency.NewEuro(-200.0), Description: "Rent"}, {Time: "4", Amount: currency.NewEuro(-50.0), Description: "Groceries"}, {Time: "5", Amount: currency.NewEuro(2000.0), Description: "Income"}}
	report := report.NewBasicBudgetReport("Test", txs)
	expected := map[string]currency.Euro{"Groceries": currency.NewEuro(-75.0), "Rent": currency.NewEuro(-200.0)}
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
	report1, err := report.ReadDefaultBudgetReportFromFile(reportName, "../testdata/defaultreport.csv")
	if err != nil {
		t.Fatalf("Failed while reading report from file: %v", err)
	}
	report2, err := report.ReadDefaultBudgetReportFromFile(reportName, "../testdata/defaultreport.csv")
	if err != nil {
		t.Fatalf("Failed while reading report from file: %v", err)
	}
	actual, err := report.CombineReports(reportName, []report.Report{report1, report2})
	if err != nil {
		t.Fatal(err)
	}
	actualBasicReport, ok := actual.(report.BasicReport)
	if !ok {
		t.Fatalf("Expected basic report, got %T", actual)
	}

	testTransactions := []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(500), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: currency.NewEuro(-200), Description: "Rent"}}
	expected := report.NewBasicBudgetReport(reportName, append(testTransactions, testTransactions...))
	if actualBasicReport.Name != expected.Name {
		t.Errorf("Expected report name %s, got %s", expected.Name, actualBasicReport.Name)
	}
	if actualBasicReport.TotalIncome.Cmp(expected.TotalIncome) != 0 {
		t.Errorf("Expected total income %s, got %s", expected.TotalIncome.String(), actualBasicReport.TotalIncome.String())
	}
	if actualBasicReport.TotalExpense.Cmp(expected.TotalExpense) != 0 {
		t.Errorf("Expected total expense %s, got %s", expected.TotalExpense.String(), expected.TotalExpense.String())
	}
	if actualBasicReport.NetIncome.Cmp(expected.NetIncome) != 0 {
		t.Errorf("Expected net income %s, got %s", expected.NetIncome.String(), actualBasicReport.NetIncome.String())
	}
	if len(actualBasicReport.Transactions()) != len(expected.Transactions()) {
		t.Errorf("Expected %d transactions, got %d", len(expected.Transactions()), len(actualBasicReport.Transactions()))
	}
}

func TestSortIncomes(t *testing.T) {
	incomes := []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(10), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(100), Description: "Income"}, {Time: "3", Amount: currency.NewEuro(50), Description: "Income"}}
	expected := []transaction.BasicTransaction{incomes[1], incomes[2], incomes[0]}
	report := report.NewBasicBudgetReport("Test", incomes)
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
	firstExpense := transaction.BasicTransaction{Time: "January 2025", Amount: currency.NewEuro(-2699.99), Description: "Rent"}
	secondExpense := transaction.BasicTransaction{Time: "February 2025", Amount: currency.NewEuro(-2699.99), Description: "Rent"}
	thirdExpense := transaction.BasicTransaction{Time: "March 2025", Amount: currency.NewEuro(-2699.99), Description: "Rent"}
	expectedSortedExpenses := []transaction.BasicTransaction{firstExpense, secondExpense, thirdExpense}
	report := report.NewBasicBudgetReport("Test", []transaction.BasicTransaction{firstExpense, secondExpense, thirdExpense})
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
	expected := []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(500), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: currency.NewEuro(-200), Description: "Rent"}}
	report := report.NewBasicBudgetReport("Test", expected)
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
	transactions := []transaction.BasicTransaction{{Time: "1", Amount: currency.NewEuro(500), Description: "Income"}, {Time: "2", Amount: currency.NewEuro(-25), Description: "Groceries"}, {Time: "3", Amount: currency.NewEuro(-200), Description: "Rent"}}
	report := report.NewBasicBudgetReport("Test", transactions)
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
