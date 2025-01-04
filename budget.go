package budget

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

func Main() int {
	fmt.Println("Welcome to the budget tracker app!")

	scanner := bufio.NewScanner(os.Stdin)
	budgetPeriodName, err := ScanPeriod(scanner)
	if err != nil {
		fmt.Println("There was an error reading the name of your budget period! Please restart the program and input a valid period name. Error: ", err)
		return 1
	}
	fmt.Printf("You've selected to record your budget for the period %s. First, enter your income(s)\n", budgetPeriodName)

	incomePeriod, err := ScanIncomePeriod(scanner, budgetPeriodName)
	if err != nil {
		fmt.Println("There was an error recording your income(s)! Please restart the program and input valid incomes. Error: ", err)
		return 1
	}

	fmt.Printf("Congrats on earning %.2f in %s!\n", incomePeriod.SumIncomes(), incomePeriod.PeriodName)
	fmt.Println("Now, enter your expense(s)")
	expensePeriod, err := ScanExpensePeriod(scanner, budgetPeriodName)
	if err != nil {
		fmt.Println("There was an error recording your expense(s)! Please restart the program and input valid expenses. Error: ", err)
		return 1
	}

	fmt.Printf("After paying your expenses for the month, you were left with %.2f!\n", incomePeriod.SumIncomes()-expensePeriod.SumExpenses())
	ScanPrintExpenseReport(os.Stdout, scanner, budgetPeriodName, incomePeriod, expensePeriod)
	return 0
}

// ScanPeriod returns the user-inputted period (string), or an error if one occurred
func ScanPeriod(scanner *bufio.Scanner) (string, error) {
	fmt.Printf("For what period would you like to record your budget? ")
	if scanner.Scan() {
		return scanner.Text(), nil
	} else {
		return "", fmt.Errorf("no input provided")
	}
}

// IncomePeriod contains a representative name for the income(s) (e.g. "January 1970")
// as well as all incomes that occurred in that time
type IncomePeriod struct {
	PeriodName string
	Incomes    []Income
}

// SumIncomes returns the sum of the income amounts
func (i *IncomePeriod) SumIncomes() float64 {
	sum := 0.0
	for _, income := range i.Incomes {
		sum += income.Amount
	}
	return sum
}

// SortIncomes sorts incomes by amount in descending order
func (i *IncomePeriod) SortIncomes() {
	slices.SortFunc(i.Incomes, func(a, b Income) int {
		if a.Amount <= b.Amount {
			return 1
		} else {
			return -1
		}
	})
}

// Income represents a single reception of money, e.g. a salary or a reimbursement
type Income struct {
	Time     string
	Amount   float64
	Category string
}

// ScanIncomePeriod scans the user-inputted period and list of incomes, and returns the collected IncomePeriod
// or an error, if one occurred
func ScanIncomePeriod(scanner *bufio.Scanner, incomePeriodName string) (*IncomePeriod, error) {
	incomes, err := ScanIncomes(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading incomes! err: %w", err)
	}
	return &IncomePeriod{PeriodName: incomePeriodName, Incomes: incomes}, nil
}

// ScanIncomes accepts user-submitted information about their income(s) and returns a slice of Income structs
// or an error, if one occurred
func ScanIncomes(scanner *bufio.Scanner) ([]Income, error) {
	var incomes []Income
	var time string
	fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return incomes, fmt.Errorf("no time for income provided")
	}
	var amount float64
	var err error
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
		incomes = append(incomes, Income{Time: time, Amount: amount, Category: "Income"})
		fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return incomes, fmt.Errorf("no time for %d%s income provided", len(incomes)+1, GetNumberEnding(len(incomes)+1))
		}
	}
	return incomes, nil
}

// ExpensePeriod consists of a name corresponding to that period as well as a slice of Expense
// for all expenses occurring in that period
type ExpensePeriod struct {
	PeriodName string
	Expenses   []Expense
}

// SumExpenses returns a sum of all the expense amounts
func (e *ExpensePeriod) SumExpenses() float64 {
	sum := 0.0
	for _, expense := range e.Expenses {
		sum += expense.Amount
	}
	return sum
}

// SortExpenses sorts the Expenses based on expense amount, descending
func (e *ExpensePeriod) SortExpenses() {
	slices.SortFunc(e.Expenses, func(a, b Expense) int {
		if a.Amount <= b.Amount {
			return 1
		} else {
			return -1
		}
	})
}

// Expense represents a single expenditure, e.g. buying food or paying rent
type Expense struct {
	Time     string
	Amount   float64
	Category string
}

// ScanExpensePeriod accepts user-submitted data about their expenses for a given period,
// marshals the data into an ExpensePeriod, and returns the result, or an error if one occurred
func ScanExpensePeriod(scanner *bufio.Scanner, expensePeriodName string) (*ExpensePeriod, error) {
	expenses, err := ScanExpenses(scanner)
	if err != nil {
		return nil, fmt.Errorf("error reading expenses! err: %w", err)
	}
	return &ExpensePeriod{PeriodName: expensePeriodName, Expenses: expenses}, nil
}

// ScanExpenses acepts user-submitted data about their expenses, marshals each instance into an Expense object,
// and returns the slice of Expenses created, or an error if one occurred
func ScanExpenses(scanner *bufio.Scanner) ([]Expense, error) {
	var expenses []Expense
	var time string
	fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, GetNumberEnding(len(expenses)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return expenses, fmt.Errorf("no time provided for the expense")
	}
	var amount float64
	var category string
	var err error
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
		fmt.Printf("What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, GetNumberEnding(len(expenses)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return expenses, fmt.Errorf("no time provided for the %d%s expense", len(expenses)+1, GetNumberEnding(len(expenses)+1))
		}
	}

	return expenses, nil
}

// ScanPrintExpenseReport asks the user to input if they would like their incomes and expenses to be printed
func ScanPrintExpenseReport(w io.Writer, scanner *bufio.Scanner, budgetPeriodName string, incomePeriod *IncomePeriod, expensePeriod *ExpensePeriod) {
	fmt.Printf("Would you like your expenses printed in CSV format for your records? [y/N] ")
	var shouldPrint string
	if scanner.Scan() {
		shouldPrint = scanner.Text()
	} else {
		fmt.Println("Error reading if you would like your expenses printed! Defaulting to No")
		return
	}
	if strings.ToLower(shouldPrint) == "y" {
		PrintExpenseReport(w, budgetPeriodName, incomePeriod, expensePeriod)
	}
}

// PrintExpenseReport prints the user's income and expenses for a given period, in CSV format
func PrintExpenseReport(w io.Writer, budgetPeriodName string, incomePeriod *IncomePeriod, expensePeriod *ExpensePeriod) {
	fmt.Fprintf(w, "Income and Expense report for the income/expense period %s\n", budgetPeriodName)
	csvWriter := csv.NewWriter(w)
	csvWriter.Write([]string{"Date", "Amount", "Category"})
	incomePeriod.SortIncomes()
	for _, income := range incomePeriod.Incomes {
		csvWriter.Write([]string{income.Time, strconv.FormatFloat(income.Amount, 'f', 2, 64), income.Category})
	}
	expensePeriod.SortExpenses()
	for _, expense := range expensePeriod.Expenses {
		csvWriter.Write([]string{expense.Time, strconv.FormatFloat(expense.Amount, 'f', 2, 64), expense.Category})
	}
	csvWriter.Flush()
}

// GetNumberEnding finds the appropriate number ending (e.g. 1st, 2nd, 3rd, ...) for a given integer
func GetNumberEnding(num int) string {
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
