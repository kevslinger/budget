package report

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kevslinger/budget/currency"
	"github.com/kevslinger/budget/transaction"
)

// BasicReport contains the information to summarize a budget
type BasicReport struct {
	Name         string
	NetIncome    currency.Euro
	TotalIncome  currency.Euro
	TotalExpense currency.Euro
	transactions []transaction.BasicTransaction
}

// NewBasicBudgetReport creates a new report with a given reportName, calculating the total income, total expense, and net income
func NewBasicBudgetReport(reportName string, transactions []transaction.BasicTransaction) BasicReport {
	totalIncome := currency.NewEuro(0.0)
	totalExpense := currency.NewEuro(0.0)
	for _, transaction := range transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) < 0 {
			totalExpense = currency.AddEuros(totalExpense, transaction.Amount)
		} else {
			totalIncome = currency.AddEuros(totalIncome, transaction.Amount)
		}
	}
	return BasicReport{Name: reportName, NetIncome: currency.AddEuros(totalIncome, totalExpense), TotalIncome: totalIncome, TotalExpense: totalExpense, transactions: transactions}
}

// ReadDefaultBudgetReportFromFile reads in a CSV file with transactions, and parses them to create a report
func ReadDefaultBudgetReportFromFile(reportName string, path string) (BasicReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return BasicReport{}, fmt.Errorf("error opening budget report file: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	line, err := reader.Read()
	if err != nil {
		return BasicReport{}, fmt.Errorf("error reading budet report file: %w", err)
	}
	var transactions []transaction.BasicTransaction
	for err == nil {
		time, amount, description := line[0], line[1], line[2]
		// skip header
		if strings.ToLower(time) == "time" && strings.ToLower(amount) == "amount" && strings.ToLower(description) == "description" {
			line, err = reader.Read()
			continue
		}
		amountFloat, parseErr := strconv.ParseFloat(amount, 64)
		if parseErr != nil {
			return BasicReport{}, fmt.Errorf("error parsing a transaction: %w", parseErr)
		}
		transactions = append(transactions, transaction.BasicTransaction{Time: time, Amount: currency.NewEuro(amountFloat), Description: description})
		line, err = reader.Read()
	}
	if !errors.Is(err, io.EOF) {
		return BasicReport{}, fmt.Errorf("error reading currency report file: %w", err)
	}

	return NewBasicBudgetReport(reportName, transactions), nil
}

func CombineBasicReports(reportName string, reports []Report) (BasicReport, error) {
	transactions := make([]transaction.BasicTransaction, 0)
	for _, r := range reports {
		basicReport, ok := r.(BasicReport)
		if !ok {
			return BasicReport{}, fmt.Errorf("expected  BasicReport, got %T", r)
		}
		transactions = append(transactions, basicReport.Transactions()...)
	}
	return NewBasicBudgetReport(reportName, transactions), nil
}

// CalculateTotalExpensePerDescription aggregates all expenses by category, and returns this data as a map
func (r BasicReport) CalculateTotalExpensePerDescription() map[string]currency.Euro {
	expensePerDescription := make(map[string]currency.Euro)
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) < 0 {
			if _, ok := expensePerDescription[transaction.Description]; !ok {
				expensePerDescription[transaction.Description] = currency.NewEuro(0.0)
			}
			expensePerDescription[transaction.Description] = currency.AddEuros(expensePerDescription[transaction.Description], transaction.Amount)
		}
	}
	return expensePerDescription
}

// Save saves the report's transactions to a CSV file
// The transasctions are saved in order:
// 1.) Incomes (sorted from largest to smallest)
// 2.) Expenses (sorted from most to least expensive)
func (r BasicReport) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = r.WriteCSV(file)
	if err != nil {
		return err
	}
	return file.Sync()
}

// WriteCSV writes the report to a CSV
func (r BasicReport) WriteCSV(writer io.Writer) error {
	fmt.Fprint(writer, "Time,Amount,Description")
	for _, income := range r.SortIncomes() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s", income.Time, float64(income.Amount.Cents())/100, income.Description)
	}
	for _, expense := range r.SortExpenses() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s", expense.Time, float64(expense.Amount.Cents())/100, expense.Description)
	}
	return nil
}

// SortIncomes sort the incomes in the report from largest to smallest
func (r BasicReport) SortIncomes() []transaction.BasicTransaction {
	var incomes []transaction.BasicTransaction
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) > 0 {
			incomes = append(incomes, transaction)
		}
	}
	return sortBasicTransactions(incomes)
}

// SortExpenses sorts expenses in the report from largest expense to smallest expense
func (r BasicReport) SortExpenses() []transaction.BasicTransaction {
	var expenses []transaction.BasicTransaction
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) < 0 {
			expenses = append(expenses, transaction)
		}
	}
	return sortBasicTransactions(expenses)
}

func sortBasicTransactions(transactions []transaction.BasicTransaction) []transaction.BasicTransaction {
	slices.SortFunc(transactions, func(a, b transaction.BasicTransaction) int {
		if math.Abs(float64(a.Amount.Cents())) <= math.Abs(float64(b.Amount.Cents())) {
			return 1
		} else {
			return -1
		}
	})
	return transactions
}

// String returns a summary of the report, including Name, total income, total expense, net income, and list of transactions
func (r BasicReport) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("currency Report for the period %s\n", r.Name))
	str.WriteString(fmt.Sprintf("Total Income: %s, Total Expense: %s, Net Income: %s\n", r.TotalIncome.String(), r.TotalExpense.String(), r.NetIncome.String()))
	str.WriteString("Total Expense Per Category (% of total expenses)\n")
	expensePerDescription := r.CalculateTotalExpensePerDescription()
	sortedKeys := sortKeys(maps.Keys(expensePerDescription))
	for _, key := range sortedKeys {
		str.WriteString(fmt.Sprintf("%s: %s (%.2f%%)\n", key, expensePerDescription[key].String(), 100*float64(expensePerDescription[key].Cents())/float64(r.TotalExpense.Cents())))
	}
	str.WriteString("Time,Amount,Description\n")
	for _, income := range r.SortIncomes() {
		str.WriteString(fmt.Sprintf("%s,%s,%s\n", income.Time, income.Amount.String(), income.Description))
	}
	for _, expense := range r.SortExpenses() {
		str.WriteString(fmt.Sprintf("%s,%s,%s\n", expense.Time, expense.Amount.String(), expense.Description))
	}
	return str.String()
}

// Transactions returns a copy of the transactions from the report
func (r BasicReport) Transactions() []transaction.BasicTransaction {
	var transactions []transaction.BasicTransaction
	for _, tx := range r.transactions {
		transactions = append(transactions, transaction.BasicTransaction{Time: tx.Time, Amount: tx.Amount, Description: tx.Description})
	}
	return transactions
}
