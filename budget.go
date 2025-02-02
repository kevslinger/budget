package budget

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kevslinger/budget/currency"
	"github.com/kevslinger/budget/report"
	"github.com/kevslinger/budget/transaction"
)

func Main() int {
	fmt.Println("Welcome to the budget tracker app! You may input budget report files as well as individual incomes and/or expenses, which will be compiled into a single report which can be saved to a CSV and/or printed to the screen.")

	scanner := bufio.NewScanner(os.Stdin)
	reportName, err := ScanPeriod(os.Stdout, scanner)
	if err != nil {
		fmt.Println("There was an error reading the name of your budget period! Please restart the program and input a valid period name. Error: ", err)
		return 1
	}
	fmt.Printf("You've selected to record your budget for the period %s.\n", reportName)

	paths := ScanReportPaths(os.Stdout, scanner)
	var reports []report.Report
	for _, path := range paths {
		report, err := report.ReadBudgetReportFromFile(reportName, path)
		if err != nil {
			fmt.Printf("There was an error reading the budget report file with path %s! Skipping this report. Error: %v", path, err)
			continue
		}
		reports = append(reports, report)
	}
	combinedReport, err := report.CombineReports(reportName, reports)
	if err != nil {
		fmt.Printf("There was an error combining the provided reports together! Please ensure the reports are compatible. Error: %v", err)
		return 1
	}
	ScanPrintExpenseReport(os.Stdout, scanner, combinedReport)
	err = ScanSaveExpenseReport(os.Stdout, scanner, combinedReport, reportName)
	if err != nil {
		fmt.Println("there was an error saving your report to CSV! Try again. Error: ", err)
		return 1
	}
	return 0
}

// ScanPeriod returns the user-inputted period (string), or an error if one occurred
func ScanPeriod(w io.Writer, scanner *bufio.Scanner) (string, error) {
	fmt.Fprint(w, "For what period would you like to record your budget? ")
	if scanner.Scan() {
		return scanner.Text(), nil
	} else {
		return "", fmt.Errorf("no input provided")
	}
}

// ScanReportPath returns the user-inputted path to the CSV file they would like to use as their report
func ScanReportPaths(w io.Writer, scanner *bufio.Scanner) []string {
	var paths []string
	fmt.Fprintf(w, "What is the path to the %d%s report CSV file? Enter -1 to stop adding paths: ", len(paths)+1, GetNumberEnding(len(paths)+1))
	for scanner.Scan() {
		path := scanner.Text()
		if path == "-1" {
			break
		}
		if len(path) == 0 {
			continue
		}
		paths = append(paths, path)
		fmt.Fprintf(w, "What is the path to the %d%s report CSV file? Enter -1 to stop adding paths: ", len(paths)+1, GetNumberEnding(len(paths)+1))
	}
	return paths
}

// ScanIncomes accepts user-submitted information about their income(s) and returns a slice of Income structs
// or an error, if one occurred
func ScanIncomes(w io.Writer, scanner *bufio.Scanner) ([]transaction.BasicTransaction, error) {
	var incomes []transaction.BasicTransaction
	var time string
	fmt.Fprintf(w, "When did you earn the %d%s income? Press -1 to stop adding incomes: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return incomes, fmt.Errorf("no time for income provided")
	}
	var amount float64
	var err error
	for time != "-1" {
		fmt.Fprintf(w, "How much did you earn? Enter the amount without currency symbol: ")
		if scanner.Scan() {
			amount, err = strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Fprintln(w, "Error reading input! Please try again. err: ", err)
				continue
			}
		} else {
			return incomes, fmt.Errorf("no amount provided for income on %s", time)
		}
		incomes = append(incomes, transaction.BasicTransaction{Time: time, Amount: currency.NewEuro(amount), Description: "Income"})
		fmt.Fprintf(w, "When did you earn the %d%s income? Press -1 to exit: ", len(incomes)+1, GetNumberEnding(len(incomes)+1))
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
func ScanExpenses(w io.Writer, scanner *bufio.Scanner) ([]transaction.BasicTransaction, error) {
	var expenses []transaction.BasicTransaction
	var time string
	fmt.Fprintf(w, "What time did the %d%s expense occur? Press -1 to stop adding expenses: ", len(expenses)+1, GetNumberEnding(len(expenses)+1))
	if scanner.Scan() {
		time = scanner.Text()
	} else {
		return expenses, fmt.Errorf("no time provided for the expense")
	}
	var amount float64
	var category string
	var err error
	for time != "-1" {
		fmt.Fprintf(w, "How much did the expense cost? Enter the amount without currency symbol: ")
		if scanner.Scan() {
			amount, err = strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Fprintln(w, "Error reading input! Please try again. err: ", err)
				continue
			}
		} else {
			return expenses, fmt.Errorf("no amount provided for expense on %s", time)
		}
		fmt.Fprintf(w, "To which category does this expense belong? ")
		if scanner.Scan() {
			category = scanner.Text()
		} else {
			return expenses, fmt.Errorf("no category provided for expense on %s of amount %.2f", time, amount)
		}
		expenses = append(expenses, transaction.BasicTransaction{Time: time, Amount: currency.NewEuro(-amount), Description: category})
		fmt.Fprintf(w, "What time did the %d%s expense occur? Press -1 to exit: ", len(expenses)+1, GetNumberEnding(len(expenses)+1))
		if scanner.Scan() {
			time = scanner.Text()
		} else {
			return expenses, fmt.Errorf("no time provided for the %d%s expense", len(expenses)+1, GetNumberEnding(len(expenses)+1))
		}
	}

	return expenses, nil
}

// ScanPrintExpenseReport asks the user to input if they would like their incomes and expenses to be printed
func ScanPrintExpenseReport(w io.Writer, scanner *bufio.Scanner, report report.Report) {
	fmt.Fprintf(w, "Would you like your report printed in CSV format for your records? [Y/n] ")
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
func PrintExpenseReport(w io.Writer, report report.Report) {
	fmt.Fprint(w, report.String())
}

// ScanSaveExpenseReport saves the user's report to a CSV file, if they wish
func ScanSaveExpenseReport(w io.Writer, scanner *bufio.Scanner, report report.Report, reportName string) error {
	shouldSave := ScanShouldSaveExpenseReport(w, scanner)
	if strings.ToLower(shouldSave) != "y" {
		return nil
	}
	reportFilename := ScanReportFilename(w, scanner, reportName)
	return report.Save(reportFilename)
}

func ScanShouldSaveExpenseReport(w io.Writer, scanner *bufio.Scanner) string {
	fmt.Fprintf(w, "Would you like your report saved to a CSV file? [Y/n]")
	y := "y"
	shouldSave := y
	if scanner.Scan() {
		shouldSave = scanner.Text()
		if len(shouldSave) == 0 {
			shouldSave = y
		}
	}
	return shouldSave
}

// ScanReportFilename gets a user-specified path to save the report
func ScanReportFilename(w io.Writer, scanner *bufio.Scanner, reportName string) string {
	fmt.Fprintf(w, "Enter the filename to use for the report. (Default: %s): ", reportName)
	filename := reportName
	if scanner.Scan() {
		filename = scanner.Text()
		if len(filename) == 0 {
			filename = reportName
		}
	}
	if !strings.HasSuffix(filename, ".csv") {
		filename = filename + ".csv"
	}
	return filename
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
