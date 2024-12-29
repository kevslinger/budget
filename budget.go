package budget

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

func Main() int {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to the budget tracker app! First, let's talk about your income.")
	incomePeriod, err := ScanIncomePeriod(scanner)
	if err != nil {
		fmt.Println("There was an error recording your income(s)! Please restart the program and input valid incomes. Error: ", err)
		return 1
	}
	totalIncome := incomePeriod.SumIncomes()
	fmt.Printf("Congrats on earning %.2f in %s!\n", totalIncome, incomePeriod.PeriodName)
	fmt.Println("Now it is time to enter your expenses.")
	expensePeriod, err := ScanExpensePeriod(scanner)
	if err != nil {
		fmt.Println("There was an error recording your expense(s)! Please restart the program and input valid expenses. Error: ", err)
		return 1
	}

	fmt.Printf("After paying your expenses for the month, you were left with %.2f!\n", incomePeriod.SumIncomes()-expensePeriod.SumExpenses())
	PrintExpenseReport(scanner, incomePeriod, expensePeriod)
	return 0
}

func ScanPeriod(scanner *bufio.Scanner) (string, error) {
	fmt.Printf("For what period would you like to record? ")
	if scanner.Scan() {
		return scanner.Text(), nil
	} else {
		return "", fmt.Errorf("no input provided")
	}
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

func (i *IncomePeriod) SortIncomes() {
	slices.SortFunc(i.Incomes, func(a, b Income) int {
		if a.Amount < b.Amount {
			return -1
		} else {
			return 1
		}
	})
}

type Income struct {
	Time     string  // TODO: time.Time? time.Date?
	Amount   float64 // TODO: big.Float?
	Category string
}

func ScanIncomePeriod(scanner *bufio.Scanner) (*IncomePeriod, error) {
	incomePeriodName, err := ScanPeriod(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading incomes! err: %w", err)
	}
	incomes, err := ScanIncomes(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading incomes! err: %w", err)
	}
	return &IncomePeriod{PeriodName: incomePeriodName, Incomes: incomes}, nil
}

func ScanIncomes(scanner *bufio.Scanner) ([]Income, error) {
	var incomes []Income
	var time string
	fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, getNumberEnding(len(incomes)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return incomes, fmt.Errorf("no time for income provided")
	}
	var amount float64
	var err error
	// TODO: Better way to end?
	for time != "-1" {
		fmt.Printf("How much did you earn? Enter the amount without currency symbol: ")
		if scanner.Scan() {
			amount, err = strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Println("Error reading input! Please try again. err: ", err)
				continue
			}
		} else {
			return incomes, fmt.Errorf("no amount provided for income on %s", time)
		}
		incomes = append(incomes, Income{Time: time, Amount: amount})
		fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, getNumberEnding(len(incomes)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return incomes, fmt.Errorf("no time for %d%s income provided", len(incomes)+1, getNumberEnding(len(incomes)+1))
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

func (e *ExpensePeriod) SortExpenses() {
	slices.SortFunc(e.Expenses, func(a, b Expense) int {
		if a.Amount < b.Amount {
			return 1
		} else {
			return -1
		}
	})
}

type Expense struct {
	Time     string  // TODO: time.Time? time.Date?
	Amount   float64 // TODO: big.Float?
	Category string
}

func ScanExpensePeriod(scanner *bufio.Scanner) (*ExpensePeriod, error) {
	expensePeriodName, err := ScanPeriod(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading expenses! err: %w", err)
	}
	expenses, err := ScanExpenses(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading expenses! err: %w", err)
	}
	return &ExpensePeriod{PeriodName: expensePeriodName, Expenses: expenses}, nil
}

func ScanExpenses(scanner *bufio.Scanner) ([]Expense, error) {
	var expenses []Expense
	var time string
	fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, getNumberEnding(len(expenses)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return expenses, fmt.Errorf("no time provided for the expense")
	}
	var amount float64
	var category string
	var err error
	// TODO: Better way to end?
	for time != "-1" {
		fmt.Printf("How much did the expense cost? Enter the amount without currency symbol: ")
		if scanner.Scan() {
			amount, err = strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Println("Error reading input! Please try again. err: ", err)
				continue
			}
		} else {
			return expenses, fmt.Errorf("no amount provided for expense on %s", time)
		}
		fmt.Printf("To which category does this expense belong? ")
		if scanner.Scan() {
			category = scanner.Text()
		} else {
			return expenses, fmt.Errorf("no category provided for expense on %s of amount %.2f", time, amount)
		}
		expenses = append(expenses, Expense{Time: time, Amount: amount, Category: category})
		fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, getNumberEnding(len(expenses)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return expenses, fmt.Errorf("no time provided for the %d%s expense", len(expenses)+1, getNumberEnding(len(expenses)+1))
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

func PrintExpenseReport(scanner *bufio.Scanner, incomePeriod *IncomePeriod, expensePeriod *ExpensePeriod) {
	fmt.Printf("Would you like your expenses printed in CSV format for your records? [y/N] ")
	var shouldPrint string
	if scanner.Scan() {
		shouldPrint = scanner.Text()
	} else {
		fmt.Println("Error reading if you would like your expenses printed! Defaulting to No")
		return
	}
	if strings.ToLower(shouldPrint) == "y" {
		fmt.Printf("Income and Expense report for the income period %s and expense period %s\n", incomePeriod.PeriodName, expensePeriod.PeriodName)
		csvWriter := csv.NewWriter(os.Stdout)
		csvWriter.Write([]string{"Date", "Amount", "Category"})
		incomePeriod.SortIncomes()
		for _, income := range incomePeriod.Incomes {
			csvWriter.Write([]string{income.Time, strconv.FormatFloat(income.Amount, 'f', 2, 64), income.Category})
		}
		expensePeriod.SortExpenses()
		for _, expense := range expensePeriod.Expenses {
			csvWriter.Write([]string{expense.Time, strconv.FormatFloat(-1*expense.Amount, 'f', 2, 64), expense.Category})
		}
		csvWriter.Flush()
	}
}
