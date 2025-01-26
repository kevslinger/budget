package budget

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Transaction contains the information to describe a single income or expense
type Transaction struct {
	Amount      Euro
	Description string
	Time        string
}

// Report contains the information to summarize a budget
type Report struct {
	Name         string
	NetIncome    Euro
	TotalIncome  Euro
	TotalExpense Euro
	transactions []Transaction
}

// NewBudgetReport creates a new report with a given reportName, calculating the total income, total expense, and net income
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

// ReadBudgetReportFromFile reads in a CSV file with transactions, and parses them to create a report
func ReadBudgetReportFromFile(reportName string, path string) (Report, error) {
	file, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("error opening budget report file: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	line, err := reader.Read()
	if err != nil {
		return Report{}, fmt.Errorf("error reading budget report file: %w", err)
	}
	var transactions []Transaction
	for err == nil {
		time, amount, description := line[0], line[1], line[2]
		// skip header
		if strings.ToLower(time) == "time" && strings.ToLower(amount) == "amount" && strings.ToLower(description) == "description" {
			line, err = reader.Read()
			continue
		}
		amountFloat, parseErr := strconv.ParseFloat(amount, 64)
		if parseErr != nil {
			return Report{}, fmt.Errorf("error parsing a transaction: %w", parseErr)
		}
		transactions = append(transactions, Transaction{Time: time, Amount: NewEuro(amountFloat), Description: description})
		line, err = reader.Read()
	}
	if !errors.Is(err, io.EOF) {
		return Report{}, fmt.Errorf("error reading budget report file: %w", err)
	}

	return NewBudgetReport(reportName, transactions), nil
}

// CombineReports copies the transactions from each report to form a new, combined report
// Ask about ... as an argument
func CombineReports(budgetPeriodName string, reports []Report) Report {
	var transactions []Transaction
	for _, report := range reports {
		transactions = append(transactions, report.Transactions()...)
	}
	return NewBudgetReport(budgetPeriodName, transactions)
}

// CalculateTotalExpensePerDescription aggregates all expenses by category, and returns this data as a map
func (r Report) CalculateTotalExpensePerDescription() map[string]Euro {
	expensePerDescription := make(map[string]Euro)
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(NewEuro(0)) < 0 {
			if _, ok := expensePerDescription[transaction.Description]; !ok {
				expensePerDescription[transaction.Description] = NewEuro(0)
			}
			expensePerDescription[transaction.Description] = AddEuros(expensePerDescription[transaction.Description], transaction.Amount)
		}
	}
	return expensePerDescription
}

// Save saves the report's transactions to a CSV file
// The transasctions are saved in order:
// 1.) Incomes (sorted from largest to smallest)
// 2.) Expenses (sorted from most to least expensive)
func (r Report) Save(filename string) error {
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

func (r Report) WriteCSV(writer io.Writer) error {
	fmt.Fprint(writer, "Time,Amount,Description")
	for _, income := range r.SortIncomes() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s", income.Time, float64(income.Amount.cents)/100, income.Description)
	}
	for _, expense := range r.SortExpenses() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s", expense.Time, float64(expense.Amount.cents)/100, expense.Description)
	}
	return nil
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
		if math.Abs(float64(a.Amount.cents)) <= math.Abs(float64(b.Amount.cents)) {
			return 1
		} else {
			return -1
		}
	})
	return transactions
}

// String returns a summary of the report, including Name, total income, total expense, net income, and list of transactions
func (r Report) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("Budget Report for the period %s\n", r.Name))
	str.WriteString(fmt.Sprintf("Total Income: %s, Total Expense: %s, Net Income: %s\n", r.TotalIncome.String(), r.TotalExpense.String(), r.NetIncome.String()))
	str.WriteString("Total Expense Per Category (% of total expenses)\n")
	expensePerDescription := r.CalculateTotalExpensePerDescription()
	sortedKeys := sortKeys(expensePerDescription)
	for _, key := range sortedKeys {
		str.WriteString(fmt.Sprintf("%s: %s (%.2f%%)\n", key, expensePerDescription[key].String(), 100*float64(expensePerDescription[key].cents)/float64(r.TotalExpense.cents)))
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
func (r Report) Transactions() []Transaction {
	var transactions []Transaction
	for _, tx := range r.transactions {
		transactions = append(transactions, Transaction{Time: tx.Time, Amount: tx.Amount, Description: tx.Description})
	}
	return transactions
}

func sortKeys(m map[string]Euro) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
