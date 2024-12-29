package budget

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Main() int {
	fmt.Println("Welcome to the budget tracker app! First, let's talk about your income.")
	incomePeriod, err := ScanIncomePeriod()
	if err != nil {
		fmt.Println("There was an error recording your income(s)! Please restart the program and input valid incomes. Error: ", err)
		return 1
	}
	totalIncome := incomePeriod.SumIncomes()
	fmt.Printf("Congrats on earning %.2ff in %s!\n", totalIncome, incomePeriod.PeriodName)
	fmt.Println("Now it is time to enter your expenses.")
	expensePeriod, err := ScanExpensePeriod()
	if err != nil {
		fmt.Println("There was an error recording your expense(s)! Please restart the program and input valid expenses. Error: ", err)
		return 1
	}

	fmt.Printf("After paying your expenses for the month, you were left with %.2f!\n", incomePeriod.SumIncomes()-expensePeriod.SumExpenses())
	PrintExpenseReport(incomePeriod, expensePeriod)
	return 0
}

func ScanPeriod() (string, error) {
	fmt.Printf("For what period would you like to record? ")
	var period string
	_, err := fmt.Scanln(&period)
	if err != nil {
		return "", fmt.Errorf("error reading the period provided! err: %w", err)
	}
	return period, nil
}

type IncomePeriod struct {
	PeriodName string
	Incomes    []Income
}

func (i *IncomePeriod) SumIncomes() float64 {
	sum := 0.0
	for _, income := range i.Incomes {
		sum += income.Amount
	}
	return sum
}

type Income struct {
	Time     string  // TODO: time.Time? time.Date?
	Amount   float64 // TODO: big.Float?
	Category string
}

func ScanIncomePeriod() (*IncomePeriod, error) {
	incomePeriodName, err := ScanPeriod()
	if err != nil {
		return nil, fmt.Errorf("error reading incomes! err: %w", err)
	}
	incomes, err := ScanIncomes()
	if err != nil {
		return nil, fmt.Errorf("error reading incomes! err: %w", err)
	}
	return &IncomePeriod{PeriodName: incomePeriodName, Incomes: incomes}, nil
}

func ScanIncomes() ([]Income, error) {
	var incomes []Income
	var time string
	fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, getNumberEnding(len(incomes)+1))
	_, err := fmt.Scanln(&time)
	if err != nil {
		return incomes, fmt.Errorf("error reading the time an income occurred! err: %w", err)
	}
	var amount float64
	// TODO: Better way to end?
	for time != "-1" {
		fmt.Printf("How much did you earn? Enter the amount without currency symbol: ")
		_, err := fmt.Scanln(&amount)
		// TODO: Continue on error?
		if err != nil {
			return incomes, fmt.Errorf("error reading the amount of an income! err: %w", err)
		}
		incomes = append(incomes, Income{Time: time, Amount: amount})
		fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, getNumberEnding(len(incomes)+1))
		_, err = fmt.Scanln(&time)
		if err != nil {
			return incomes, fmt.Errorf("error reading the time of an income occurred! err: %w", err)
		}
	}

	return incomes, nil
}

// TODO: Naming?
type ExpensePeriod struct {
	PeriodName string
	Expenses   []Expense
}

func (e *ExpensePeriod) SumExpenses() float64 {
	sum := 0.0
	for _, expense := range e.Expenses {
		sum += expense.Amount
	}
	return sum
}

type Expense struct {
	Time     string  // TODO: time.Time? time.Date?
	Amount   float64 // TODO: big.Float?
	Category string
}

func ScanExpensePeriod() (*ExpensePeriod, error) {
	expensePeriodName, err := ScanPeriod()
	if err != nil {
		return nil, fmt.Errorf("error reading expenses! err: %w", err)
	}
	expenses, err := ScanExpenses()
	if err != nil {
		return nil, fmt.Errorf("error reading expenses! err: %w", err)
	}
	return &ExpensePeriod{PeriodName: expensePeriodName, Expenses: expenses}, nil
}

func ScanExpenses() ([]Expense, error) {
	var expenses []Expense
	var time string
	fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, getNumberEnding(len(expenses)+1))
	_, err := fmt.Scanln(&time)
	if err != nil {
		return expenses, fmt.Errorf("error reading the time an expense occurred! err: %w", err)
	}
	var amount float64
	var category string
	// TODO: Better way to end?
	for time != "-1" {
		fmt.Printf("How much did the expense cost? Enter the amount without currency symbol: ")
		_, err := fmt.Scanln(&amount)
		// TODO: Continue on error?
		if err != nil {
			return expenses, fmt.Errorf("error reading the amount of an expense! err: %w", err)
		}
		fmt.Printf("To which category does this expense belong? ")
		_, err = fmt.Scanln(&category)
		// TODO: Continue on error?
		if err != nil {
			return expenses, fmt.Errorf("error reading the category of an expense! err: %w", err)
		}
		expenses = append(expenses, Expense{Time: time, Amount: amount, Category: category})
		fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, getNumberEnding(len(expenses)+1))
		_, err = fmt.Scanln(&time)
		if err != nil {
			return expenses, fmt.Errorf("error reading the time an expense occurred! err: %w", err)
		}
	}

	return expenses, nil
}

func getNumberEnding(num int) string {
	switch num % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func PrintExpenseReport(incomePeriod *IncomePeriod, expensePeriod *ExpensePeriod) {
	fmt.Printf("Would you like your expenses printed in CSV format for your records? [y/N] ")
	var shouldPrint string
	_, err := fmt.Scanln(&shouldPrint)
	if err != nil {
		fmt.Println("Error reading if you would like your expenses printed! Defaulting to No")
	}
	if strings.ToLower(shouldPrint) == "y" {
		fmt.Printf("Income and Expense report for the income period %s and expense period %s\n", incomePeriod.PeriodName, expensePeriod.PeriodName)
		csvWriter := csv.NewWriter(os.Stdout)
		csvWriter.Write([]string{"Date", "Amount", "Category"})
		for _, income := range incomePeriod.Incomes {
			csvWriter.Write([]string{income.Time, strconv.FormatFloat(income.Amount, 'f', 2, 64), income.Category})
		}
		for _, expense := range expensePeriod.Expenses {
			csvWriter.Write([]string{expense.Time, strconv.FormatFloat(expense.Amount, 'f', 2, 64), expense.Category})
		}
		csvWriter.Flush()
	}
}
