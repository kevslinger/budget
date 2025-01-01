package budget_test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevslinger/budget"
)

func TestScanPeriodReturnsExpectedResult(t *testing.T) {
	expectedPeriodNames := []string{"January 2025", "2024", "January through June 2024"}
	for _, expectedPeriodName := range expectedPeriodNames {
		actualPeriodName, err := budget.ScanPeriod(bufio.NewScanner(strings.NewReader(expectedPeriodName)))
		if err != nil {
			t.Fatalf("Got error reading period name: %v", err)
		}
		if actualPeriodName != expectedPeriodName {
			t.Fatalf("Expected %s but got %s", expectedPeriodName, actualPeriodName)
		}
	}
}

func TestScanIncomePeriod(t *testing.T) {
	incomePeriodName := "Testing Scan Income Period"
	firstIncomeTime := "January 5"
	firstIncomeAmount := 500
	secondIncomeTime := "January 6"
	secondIncomeAmount := 1000
	thirdIncomeTime := "February 1"
	thirdIncomeAmount := 200

	incomeScanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf("%s\n%d\n%s\n%d\n%s\n%d\n-1", firstIncomeTime, firstIncomeAmount, secondIncomeTime, secondIncomeAmount, thirdIncomeTime, thirdIncomeAmount)))
	expectedIncomePeriod := &budget.IncomePeriod{PeriodName: incomePeriodName, Incomes: []budget.Income{{Time: firstIncomeTime, Amount: float64(firstIncomeAmount), Category: "Income"}, {Time: secondIncomeTime, Amount: float64(secondIncomeAmount), Category: "Income"}, {Time: thirdIncomeTime, Amount: float64(thirdIncomeAmount), Category: "Income"}}}
	actualIncomePeriod, err := budget.ScanIncomePeriod(incomeScanner, incomePeriodName)
	if err != nil {
		t.Fatalf("Got error reading income period: %v", err)
	}
	if !cmp.Equal(expectedIncomePeriod, actualIncomePeriod) {
		t.Fatalf("Expected %#v but got %#v", expectedIncomePeriod, actualIncomePeriod)
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
	expectedIncomes := []budget.Income{{Time: firstIncomeTime, Amount: firstIncomeAmount, Category: "Income"}, {Time: secondIncomeTime, Amount: secondIncomeAmount, Category: "Income"}, {Time: thirdIncomeTime, Amount: thirdIncomeAmount, Category: "Income"}}
	actualIncomes, err := budget.ScanIncomes(incomeScanner)
	if err != nil {
		t.Fatalf("Got error reading incomes: %v", err)
	}
	if !cmp.Equal(expectedIncomes, actualIncomes) {
		t.Fatalf("Expected %#v but got %#v", expectedIncomes, actualIncomes)
	}
}

func TestScanExpensePeriod(t *testing.T) {
	expensePeriodName := "Testing Scan Expense Period"
	firstExpenseTime := "March 1"
	firstExpenseAmount := 500.99
	firstExpenseCategory := "Groceries"
	secondExpenseTime := "March 28"
	secondExpenseAmount := 1000.00
	secondExpenseCategory := "Rent"
	thirdExpenseTime := "March 30"
	thirdExpenseAmount := 200.99
	thirdExpenseCategory := "Travel"

	expensePeriodScanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf("%s\n%.2f\n%s\n%s\n%.2f\n%s\n%s\n%.2f\n%s\n-1", firstExpenseTime, firstExpenseAmount, firstExpenseCategory, secondExpenseTime, secondExpenseAmount, secondExpenseCategory, thirdExpenseTime, thirdExpenseAmount, thirdExpenseCategory)))
	expectedExpensePeriod := &budget.ExpensePeriod{PeriodName: expensePeriodName, Expenses: []budget.Expense{{Time: firstExpenseTime, Amount: firstExpenseAmount, Category: firstExpenseCategory}, {Time: secondExpenseTime, Amount: secondExpenseAmount, Category: secondExpenseCategory}, {Time: thirdExpenseTime, Amount: thirdExpenseAmount, Category: thirdExpenseCategory}}}
	actualExpensePeriod, err := budget.ScanExpensePeriod(expensePeriodScanner, expensePeriodName)
	if err != nil {
		t.Fatalf("Got error reading expense period: %v", err)
	}
	if !cmp.Equal(expectedExpensePeriod, actualExpensePeriod) {
		t.Fatalf("Expected %#v but got %#v", expectedExpensePeriod, actualExpensePeriod)
	}
}

func TestScanExpenses(t *testing.T) {
	firstExpenseTime := "1"
	firstExpenseAmount := 0.98
	firstExpenseCategory := "Other"
	secondExpenseTime := "2"
	secondExpenseAmount := 56.95
	secondExpenseCategory := "Takeout"
	thirdExpenseTime := "5"
	thirdExpenseAmount := 3.50
	thirdExpenseCategory := "Groceries"

	expenseScanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf("%s\n%.2f\n%s\n%s\n%.2f\n%s\n%s\n%.2f\n%s\n-1", firstExpenseTime, firstExpenseAmount, firstExpenseCategory, secondExpenseTime, secondExpenseAmount, secondExpenseCategory, thirdExpenseTime, thirdExpenseAmount, thirdExpenseCategory)))
	expectedExpenses := []budget.Expense{{Time: firstExpenseTime, Amount: firstExpenseAmount, Category: firstExpenseCategory}, {Time: secondExpenseTime, Amount: secondExpenseAmount, Category: secondExpenseCategory}, {Time: thirdExpenseTime, Amount: thirdExpenseAmount, Category: thirdExpenseCategory}}
	actualExpenses, err := budget.ScanExpenses(expenseScanner)
	if err != nil {
		t.Fatalf("Got error reading expenses: %v", err)
	}
	if !cmp.Equal(expectedExpenses, actualExpenses) {
		t.Fatalf("Expected %#v but got %#v", expectedExpenses, actualExpenses)
	}
}

func TestScanPrintExpenseReportEmptyInputDefaultsToNo(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("\n"))
	budget.ScanPrintExpenseReport(nil, scanner, "", nil, nil)
}

func TestPrintExpenseReport(t *testing.T) {
	budgetPeriodName := "2024"
	incomeTime := "January"
	incomeAmount := 5000.0
	incomeCategory := "Income"
	incomePeriod := &budget.IncomePeriod{PeriodName: budgetPeriodName, Incomes: []budget.Income{{Time: incomeTime, Amount: incomeAmount, Category: incomeCategory}}}
	expenseTime := "January"
	expenseAmount := 4999.0
	expenseCategory := "Rent"
	expensePeriod := &budget.ExpensePeriod{PeriodName: budgetPeriodName, Expenses: []budget.Expense{{Time: expenseTime, Amount: expenseAmount, Category: expenseCategory}}}
	expected := fmt.Sprintf("Income and Expense report for the income/expense period %s\nDate,Amount,Category\n%s,%.2f,%s\n%s,%.2f,%s\n", budgetPeriodName, incomeTime, incomeAmount, incomeCategory, expenseTime, expenseAmount, expenseCategory)
	w := &strings.Builder{}
	budget.PrintExpenseReport(w, budgetPeriodName, incomePeriod, expensePeriod)
	actual := w.String()
	if expected != actual {
		t.Fatalf("Expected %s got %s", expected, actual)
	}
}

func TestSumIncomes(t *testing.T) {
	incomePeriod := budget.IncomePeriod{Incomes: []budget.Income{{Amount: 100.5}, {Amount: 2000}, {Amount: 0.5}}}
	expected := 2101.0
	actual := incomePeriod.SumIncomes()
	if expected != actual {
		t.Fatalf("Expected %f got %f", expected, actual)
	}
}

func TestSortIncomes(t *testing.T) {
	firstIncome := budget.Income{Time: "1", Amount: 2000.0}
	secondIncome := budget.Income{Time: "2", Amount: 675.4}
	thirdIncome := budget.Income{Time: "3", Amount: 500.1}
	sortedIncomes := []budget.Income{firstIncome, secondIncome, thirdIncome}
	incomePeriod := budget.IncomePeriod{Incomes: []budget.Income{thirdIncome, firstIncome, secondIncome}}
	incomePeriod.SortIncomes()
	if !cmp.Equal(sortedIncomes, incomePeriod.Incomes) {
		t.Fatalf("Expected %#v got %#v", sortedIncomes, incomePeriod.Incomes)
	}
}

func TestSumExpenses(t *testing.T) {
	expensePeriod := budget.ExpensePeriod{Expenses: []budget.Expense{{Amount: 500.5}, {Amount: 1000}, {Amount: 0.95}}}
	expected := 1501.45
	actual := expensePeriod.SumExpenses()
	if expected != actual {
		t.Fatalf("Expected %f but got %f", expected, actual)
	}
}

func TestSortExpenses(t *testing.T) {
	firstExpense := budget.Expense{Time: "January 2025", Amount: 2699.99, Category: "Rent"}
	secondExpense := budget.Expense{Time: "February 2025", Amount: 2699.99, Category: "Rent"}
	thirdExpense := budget.Expense{Time: "March 2025", Amount: 2699.99, Category: "Rent"}
	sortedExpense := []budget.Expense{firstExpense, secondExpense, thirdExpense}
	expensePeriod := budget.ExpensePeriod{Expenses: []budget.Expense{firstExpense, secondExpense, thirdExpense}}
	expensePeriod.SortExpenses()
	if !cmp.Equal(sortedExpense, expensePeriod.Expenses) {
		t.Fatalf("Expected %#v got %#v", sortedExpense, expensePeriod.Expenses)
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
