package budget

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func Main() int {
	fmt.Println("Welcome to the budget tracker app! You may input budget report files as well as individual incomes and/or expenses, which will be compiled into a single report which can be saved to a CSV and/or printed to the screen.")

	scanner := bufio.NewScanner(os.Stdin)
	budgetPeriodName, err := ScanPeriod(scanner)
	if err != nil {
		fmt.Println("There was an error reading the name of your budget period! Please restart the program and input a valid period name. Error: ", err)
		return 1
	}
	fmt.Printf("You've selected to record your budget for the period %s.\n", budgetPeriodName)

	paths := ScanReportPaths(scanner)
	var reports []Report
	for _, path := range paths {
		report, err := ReadBudgetReportFromFile(budgetPeriodName, path)
		if err != nil {
			fmt.Printf("There was an error reading the budget report file with path %s! Skipping this report. Error: %v", path, err)
			continue
		}
		reports = append(reports, report)
	}
	incomes, err := ScanIncomes(scanner)
	if err != nil {
		fmt.Println("There was an error recording your incomes! These will be ignored. Error: ", err)
	}
	expenses, err := ScanExpenses(scanner)
	if err != nil {
		fmt.Println("There was an error recording your expenses! These will be ignored. Error: ", err)
	}

	combinedReport := CombineReports(budgetPeriodName, append(reports, NewBudgetReport(budgetPeriodName, append(incomes, expenses...))))
	ScanPrintExpenseReport(os.Stdout, scanner, combinedReport)
	err = ScanSaveExpenseReport(scanner, combinedReport)
	if err != nil {
		fmt.Println("there was an error saving your report to CSV! Try again. Error: ", err)
		return 1
	}
	return 0
}

// ScanPeriod returns the user-inputted period (string), or an error if one occurred
func ScanPeriod(scanner *bufio.Scanner) (string, error) {
	fmt.Print("For what period would you like to record your budget? ")
	if scanner.Scan() {
		return scanner.Text(), nil
	} else {
		return "", fmt.Errorf("no input provided")
	}
}

// ScanReportPath returns the user-inputted path to the CSV file they would like to use as their report
func ScanReportPath(scanner *bufio.Scanner) (string, error) {
	fmt.Print("What is the path to the report CSV file? ")
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("no path provided")
}

// ScanReportPath returns the user-inputted path to the CSV file they would like to use as their report
func ScanReportPaths(scanner *bufio.Scanner) []string {
	var paths []string
	fmt.Printf("What is the path to the %d%s report CSV file? Enter -1 to stop adding paths: ", len(paths)+1, GetNumberEnding(len(paths)+1))
	for scanner.Scan() {
		path := scanner.Text()
		if path == "-1" {
			break
		}
		if len(path) == 0 {
			continue
		}
		paths = append(paths, path)
		fmt.Printf("What is the path to the %d%s report CSV file? Enter -1 to stop adding paths: ", len(paths)+1, GetNumberEnding(len(paths)+1))
	}
	return paths
}

// ScanIncomes accepts user-submitted information about their income(s) and returns a slice of Income structs
// or an error, if one occurred
func ScanIncomes(scanner *bufio.Scanner) ([]Transaction, error) {
	var incomes []Transaction
	var time string
	fmt.Printf("When did you earn the %d%s income? Press -1 to stop adding incomes: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
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
	fmt.Printf("What time did the %d%s expense occur? Press -1 to stop adding expenses: ", len(expenses)+1, GetNumberEnding(len(expenses)+1))
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
	fmt.Printf("Would you like your report printed in CSV format for your records? [Y/n] ")
	y := "y"
	shouldPrint := y
	if scanner.Scan() {
		shouldPrint = scanner.Text()
		if len(shouldPrint) == 0 {
			shouldPrint = y
		}
	}
	if strings.ToLower(shouldPrint) == y {
		PrintExpenseReport(w, report)
	}
}

// PrintExpenseReport prints the user's income and expenses for a given period, in CSV format
func PrintExpenseReport(w io.Writer, report Report) {
	fmt.Fprint(w, report.String())
}

// ScanSaveExpenseReport saves the user's report to a CSV file, if they wish
func ScanSaveExpenseReport(scanner *bufio.Scanner, report Report) error {
	fmt.Printf("Would you like your report saved to a CSV file? [Y/n]")
	y := "y"
	shouldSave := y
	if scanner.Scan() {
		shouldSave = scanner.Text()
		if len(shouldSave) == 0 {
			shouldSave = y
		}
	}
	if strings.ToLower(shouldSave) == y {
		return SaveExpenseReport(scanner, report)
	}
	return nil
}

// SaveExpenseReport saves the report to a user-specified path
func SaveExpenseReport(scanner *bufio.Scanner, report Report) error {
	fmt.Printf("Enter the filename to use for the report. (Default: %s): ", report.Name)
	filename := report.Name
	if scanner.Scan() {
		filename = scanner.Text()
		if len(filename) == 0 {
			filename = report.Name
		}
	}
	if !strings.HasSuffix(filename, ".csv") {
		filename = filename + ".csv"
	}
	return report.Save(filename)
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
