package budget

import (
	"bufio"
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

	incomes, err := ScanIncomes(scanner)
	if err != nil {
		fmt.Println("There was an error recording your income(s)! Please restart the program and input valid incomes. Error: ", err)
		return 1
	}

	fmt.Println("Now, enter your expense(s)")
	expenses, err := ScanExpenses(scanner)
	if err != nil {
		fmt.Println("There was an error recording your expense(s)! Please restart the program and input valid expenses. Error: ", err)
		return 1
	}
	report := NewBudgetReport(budgetPeriodName, append(incomes, expenses...))

	ScanPrintExpenseReport(os.Stdout, scanner, report)
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

type Transaction struct {
	Amount      Euro
	Description string
	Time        string
}

type Report struct {
	Name         string
	NetIncome    Euro
	TotalIncome  Euro
	TotalExpense Euro
	transactions []Transaction
}

func NewBudgetReport(reportName string, transactions []Transaction) Report {
	totalIncome := NewEuro(0.0)
	totalExpense := NewEuro(0.0)
	for _, transaction := range transactions {
		if transaction.Amount.cents < 0 {
			totalExpense = AddEuros(totalExpense, transaction.Amount)
		} else {
			totalIncome = AddEuros(totalIncome, transaction.Amount)
		}
	}
	return Report{Name: reportName, NetIncome: AddEuros(totalIncome, totalExpense), TotalIncome: totalIncome, TotalExpense: totalExpense, transactions: transactions}
}

// SortIncomes sort the incomes in the report from largest to smallest
func (r Report) SortIncomes() []Transaction {
	var incomes []Transaction
	for _, transaction := range r.transactions {
		if transaction.Amount.cents > 0 {
			incomes = append(incomes, transaction)
		}
	}
	return sortTransactions(incomes)
}

// SortExpenses sorts expenses in the report from largest expense to smallest expense
func (r Report) SortExpenses() []Transaction {
	var expenses []Transaction
	for _, transaction := range r.transactions {
		if transaction.Amount.cents < 0 {
			expenses = append(expenses, transaction)
		}
	}
	return sortTransactions(expenses)
}

func sortTransactions(transactions []Transaction) []Transaction {
	slices.SortFunc(transactions, func(a, b Transaction) int {
		if a.Amount.cents <= b.Amount.cents {
			return 1
		} else {
			return -1
		}
	})
	return transactions
}

func (r Report) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("Budget Report for the period %s\n", r.Name))
	str.WriteString(fmt.Sprintf("Total Income: %s, Total Expense: %s, Net Income: %s\n", r.TotalIncome.String(), r.TotalExpense.String(), r.NetIncome.String()))
	str.WriteString("Time,Amount,Description\n")
	for _, income := range r.SortIncomes() {
		str.WriteString(fmt.Sprintf("%s,%s,%s\n", income.Time, income.Amount.String(), income.Description))
	}
	for _, expense := range r.SortExpenses() {
		str.WriteString(fmt.Sprintf("%s,%s,%s\n", expense.Time, expense.Amount.String(), expense.Description))
	}
	return str.String()
}

// ScanIncomes accepts user-submitted information about their income(s) and returns a slice of Income structs
// or an error, if one occurred
func ScanIncomes(scanner *bufio.Scanner) ([]Transaction, error) {
	var incomes []Transaction
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
		incomes = append(incomes, Transaction{Time: time, Amount: NewEuro(amount), Description: "Income"})
		fmt.Printf("When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return incomes, fmt.Errorf("no time for %d%s income provided", len(incomes)+1, GetNumberEnding(len(incomes)+1))
		}
	}
	return incomes, nil
}

// ScanExpenses acepts user-submitted data about their expenses, marshals each instance into an Expense object,
// and returns the slice of Expenses created, or an error if one occurred
func ScanExpenses(scanner *bufio.Scanner) ([]Transaction, error) {
	var expenses []Transaction
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
		expenses = append(expenses, Transaction{Time: time, Amount: NewEuro(-amount), Description: category})
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
func ScanPrintExpenseReport(w io.Writer, scanner *bufio.Scanner, report Report) {
	fmt.Printf("Would you like your expenses printed in CSV format for your records? [y/N] ")
	var shouldPrint string
	if scanner.Scan() {
		shouldPrint = scanner.Text()
	} else {
		fmt.Println("Error reading if you would like your expenses printed! Defaulting to No")
		return
	}
	if strings.ToLower(shouldPrint) == "y" {
		PrintExpenseReport(w, report)
	}
}

// PrintExpenseReport prints the user's income and expenses for a given period, in CSV format
func PrintExpenseReport(w io.Writer, report Report) {
	fmt.Fprint(w, report.String())
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
