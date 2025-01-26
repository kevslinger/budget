package budget_test

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/kevslinger/budget"
)

func TestScanPeriodReturnsExpectedResult(t *testing.T) {
	expectedPeriodNames := []string{"January 2025", "2024", "January through June 2024"}
	for _, expectedPeriodName := range expectedPeriodNames {
		actualPeriodName, err := budget.ScanPeriod(new(bytes.Buffer), bufio.NewScanner(strings.NewReader(expectedPeriodName)))
		if err != nil {
			t.Fatalf("Got error reading period name: %v", err)
		}
		if actualPeriodName != expectedPeriodName {
			t.Fatalf("Expected %s but got %s", expectedPeriodName, actualPeriodName)
		}
	}
}

func TestScanPeriodWithoutInputReturnsError(t *testing.T) {
	got, err := budget.ScanPeriod(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("")))
	if err == nil {
		t.Errorf("Expected error, got %v", err)
	}
	if got != "" {
		t.Errorf("Expected \"\", got %s", got)
	}
}

func TestScanReportPaths(t *testing.T) {
	expected := []string{"report.csv"}
	actual := budget.ScanReportPaths(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("report.csv\n\n-1")))
	if len(expected) != len(actual) {
		t.Fatalf("Expected %d paths, got %d", len(expected), len(actual))
	}
	for idx := range expected {
		if expected[idx] != actual[idx] {
			t.Errorf("Expected %s, got %s", expected[idx], actual[idx])
		}
	}
}

func TestScanIncomes(t *testing.T) {
	firstIncomeTime := "8am"
	firstIncomeAmount := 100.0
	secondIncomeTime := "5pm"
	secondIncomeAmount := 900.50
	thirdIncomeTime := "Midnight"
	thirdIncomeAmount := 500.75

	incomeScanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf("%s\n%.2f\n%s\n%.2f\n%s\n%.2f\n-1", firstIncomeTime, firstIncomeAmount, secondIncomeTime, secondIncomeAmount, thirdIncomeTime, thirdIncomeAmount)))
	expectedIncomes := []budget.Transaction{{Time: firstIncomeTime, Amount: budget.NewEuro(firstIncomeAmount), Description: "Income"}, {Time: secondIncomeTime, Amount: budget.NewEuro(secondIncomeAmount), Description: "Income"}, {Time: thirdIncomeTime, Amount: budget.NewEuro(thirdIncomeAmount), Description: "Income"}}
	actualIncomes, err := budget.ScanIncomes(new(bytes.Buffer), incomeScanner)
	if err != nil {
		t.Fatalf("Got error reading incomes: %v", err)
	}
	if len(expectedIncomes) != len(actualIncomes) {
		t.Fatalf("Expectted %#v got %#v", expectedIncomes, actualIncomes)
	}
	for idx := range len(expectedIncomes) {
		expected, actual := expectedIncomes[idx], actualIncomes[idx]
		if expected.Time != actual.Time || expected.Amount.Cmp(actual.Amount) != 0 || expected.Description != actual.Description {
			t.Fatalf("Expected %#v got %#v", expectedIncomes, actualIncomes)
		}
	}
}

func TestScanIncomesErrorsWithEmptyInput(t *testing.T) {
	_, err := budget.ScanIncomes(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("")))
	if err == nil {
		t.Error(err)
	}
	_, err = budget.ScanIncomes(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("1\n\n")))
	if err == nil {
		t.Error(err)
	}
	_, err = budget.ScanIncomes(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("1\n500\n")))
	if err == nil {
		t.Error(err)
	}
}

func TestScanExpenses(t *testing.T) {
	firstExpenseTime := "1"
	firstExpenseAmount := 0.98
	firstExpenseDescription := "Other"
	secondExpenseTime := "2"
	secondExpenseAmount := 56.95
	secondExpenseDescription := "Takeout"
	thirdExpenseTime := "5"
	thirdExpenseAmount := 3.50
	thirdExpenseDescription := "Groceries"

	expenseScanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf("%s\n%.2f\n%s\n%s\n%.2f\n%s\n%s\n%.2f\n%s\n-1", firstExpenseTime, firstExpenseAmount, firstExpenseDescription, secondExpenseTime, secondExpenseAmount, secondExpenseDescription, thirdExpenseTime, thirdExpenseAmount, thirdExpenseDescription)))
	expectedExpenses := []budget.Transaction{{Time: firstExpenseTime, Amount: budget.NewEuro(-firstExpenseAmount), Description: firstExpenseDescription}, {Time: secondExpenseTime, Amount: budget.NewEuro(-secondExpenseAmount), Description: secondExpenseDescription}, {Time: thirdExpenseTime, Amount: budget.NewEuro(-thirdExpenseAmount), Description: thirdExpenseDescription}}
	actualExpenses, err := budget.ScanExpenses(new(bytes.Buffer), expenseScanner)
	if err != nil {
		t.Fatalf("Got error reading expenses: %v", err)
	}
	if len(expectedExpenses) != len(actualExpenses) {
		t.Fatalf("Expected %#v got %#v", expectedExpenses, actualExpenses)
	}
	for idx := range len(expectedExpenses) {
		expected, actual := expectedExpenses[idx], actualExpenses[idx]
		if expected.Time != actual.Time || expected.Amount.Cmp(actual.Amount) != 0 || expected.Description != actual.Description {
			t.Fatalf("Expected %#v got %#v", expectedExpenses, actualExpenses)
		}
	}
}

func TestScanExpensesErrorsWithEmptyInput(t *testing.T) {
	_, err := budget.ScanExpenses(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("")))
	if err == nil {
		t.Error(err)
	}
	_, err = budget.ScanExpenses(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("1\n\n")))
	if err == nil {
		t.Error(err)
	}
	_, err = budget.ScanExpenses(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("1\n500\n")))
	if err == nil {
		t.Error(err)
	}
	_, err = budget.ScanExpenses(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("1\n500\nCategory\n")))
	if err == nil {
		t.Error(err)
	}
}

func TestScanPrintExpenseReportEmptyInputDefaultsToYes(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("\n"))
	budget.ScanPrintExpenseReport(&strings.Builder{}, scanner, budget.Report{})
}

func TestPrintExpenseReport(t *testing.T) {
	budgetName := "2024"
	incomeTime := "January"
	incomeAmount := 5000.0
	incomeDescription := "Income"
	expenseTime := "January"
	expenseAmount := -4999.0
	expenseDescription := "Rent"
	report := budget.NewBudgetReport(budgetName, []budget.Transaction{{Time: incomeTime, Amount: budget.NewEuro(incomeAmount), Description: incomeDescription}, {Time: expenseTime, Amount: budget.NewEuro(expenseAmount), Description: expenseDescription}})
	expected := report.String()
	w := &strings.Builder{}
	budget.PrintExpenseReport(w, report)
	actual := w.String()
	if expected != actual {
		t.Fatalf("Expected %s got %s", expected, actual)
	}
}

func TestScanShouldSaveExpenseReport(t *testing.T) {
	expected := "y"
	actual := budget.ScanShouldSaveExpenseReport(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("")))
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

	actual = budget.ScanShouldSaveExpenseReport(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("\n")))
	if expected != strings.ToLower(actual) {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

	actual = budget.ScanShouldSaveExpenseReport(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("Y")))
	if expected != strings.ToLower(actual) {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

	expected = "n"
	actual = budget.ScanShouldSaveExpenseReport(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("n")))
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestScanReportFilename(t *testing.T) {
	expected := "report.csv"
	actual := budget.ScanReportFilename(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("report")), "test")
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
	actual = budget.ScanReportFilename(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("")), "report")
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
	actual = budget.ScanReportFilename(new(bytes.Buffer), bufio.NewScanner(strings.NewReader("\n")), "report")
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestGetNumberEnding(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 25, 33, 192, 566}
	expected := []string{"st", "nd", "rd", "th", "th", "th", "th", "th", "th", "th", "th", "rd", "nd", "th"}
	for idx, number := range numbers {
		actual := budget.GetNumberEnding(number)
		if actual != expected[idx] {
			t.Fatalf("Expected %s for %d but got %s", expected[idx], number, actual)
		}
	}
}
