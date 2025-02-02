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

// MultiPayerReport contains the information to summarize a shared Budget between multiple people
type MultiPayerReport struct {
	Name                 string
	NetIncome            currency.Euro
	TotalIncome          currency.Euro
	TotalExpense         currency.Euro
	NetIncomePerPayer    map[string]currency.Euro
	TotalIncomePerPayer  map[string]currency.Euro
	TotalExpensePerPayer map[string]currency.Euro
	transactions         []transaction.PayerTransaction
}

func NewMultiPayerBudgetReport(reportName string, transactions []transaction.PayerTransaction) MultiPayerReport {
	totalIncomePerPayer := make(map[string]currency.Euro)
	totalExpensePerPayer := make(map[string]currency.Euro)
	for _, transaction := range transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) < 0 {
			totalExpensePerPayer[transaction.PaidBy] = currency.AddEuros(totalExpensePerPayer[transaction.PaidBy], transaction.Amount)
		} else {
			totalIncomePerPayer[transaction.PaidBy] = currency.AddEuros(totalExpensePerPayer[transaction.PaidBy], transaction.Amount)
		}
	}
	netIncomePerPayer := make(map[string]currency.Euro)
	totalIncome := currency.NewEuro(0.0)
	for payer := range totalIncomePerPayer {
		netIncomePerPayer[payer] = currency.AddEuros(netIncomePerPayer[payer], totalIncomePerPayer[payer])
		totalIncome = currency.AddEuros(totalIncome, totalIncomePerPayer[payer])
	}
	totalExpense := currency.NewEuro(0.0)
	for payer := range totalExpensePerPayer {
		netIncomePerPayer[payer] = currency.AddEuros(netIncomePerPayer[payer], totalExpensePerPayer[payer])
		totalExpense = currency.AddEuros(totalExpense, totalExpensePerPayer[payer])
	}
	return MultiPayerReport{Name: reportName, NetIncome: currency.AddEuros(totalIncome, totalExpense), TotalIncome: totalIncome, TotalExpense: totalExpense, NetIncomePerPayer: netIncomePerPayer, TotalIncomePerPayer: totalIncomePerPayer, TotalExpensePerPayer: totalExpensePerPayer, transactions: transactions}
}

func ReadMultiPayerBudgetReportFromFile(reportName string, path string) (MultiPayerReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return MultiPayerReport{}, fmt.Errorf("error opening budget report file: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	line, err := reader.Read()
	if err != nil {
		return MultiPayerReport{}, fmt.Errorf("error reading budet report file: %w", err)
	}
	var transactions []transaction.PayerTransaction
	for err == nil {
		time, amount, description, payer := line[0], line[1], line[2], line[3]
		// skip header
		if strings.ToLower(time) == "time" && strings.ToLower(amount) == "amount" && strings.ToLower(description) == "description" {
			line, err = reader.Read()
			continue
		}
		amountFloat, parseErr := strconv.ParseFloat(amount, 64)
		if parseErr != nil {
			return MultiPayerReport{}, fmt.Errorf("error parsing a transaction: %w", parseErr)
		}
		transactions = append(transactions, transaction.PayerTransaction{Time: time, Amount: currency.NewEuro(amountFloat), Description: description, PaidBy: payer})
		line, err = reader.Read()
	}
	if !errors.Is(err, io.EOF) {
		return MultiPayerReport{}, fmt.Errorf("error reading currency report file: %w", err)
	}

	return NewMultiPayerBudgetReport(reportName, transactions), nil
}

func CombineMultiPayerReports(reportName string, reports []Report) (MultiPayerReport, error) {
	transactions := make([]transaction.PayerTransaction, 0)
	for _, r := range reports {
		multiPayerReport, ok := r.(MultiPayerReport)
		if !ok {
			return MultiPayerReport{}, fmt.Errorf("expected MultiPayerReport, got %T", r)
		}
		transactions = append(transactions, multiPayerReport.Transactions()...)
	}
	return NewMultiPayerBudgetReport(reportName, transactions), nil
}

// CalculateTotalExpensePerDescription aggregates all expenses by category, and returns this data as a map
func (r MultiPayerReport) CalculateTotalExpensePerDescription() map[string]currency.Euro {
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
func (r MultiPayerReport) Save(filename string) error {
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

/*
// MultiPayerReport contains the information to summarize a shared Budget between multiple people
type MultiPayerReport struct {
	Name                 string
	NetIncome            currency.Euro
	TotalIncome          currency.Euro
	TotalExpense         currency.Euro
	NetIncomePerPayer    map[string]currency.Euro
	TotalIncomePerPayer  map[string]currency.Euro
	TotalExpensePerPayer map[string]currency.Euro
	transactions         []transaction.PayerTransaction
}

*/

// String returns a summary of the report, including the name, income/expenses per payer, and list of transactions
func (r MultiPayerReport) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("currency Report for the period %s\n", r.Name))
	str.WriteString(fmt.Sprintf("Total Income: %s, Total Expense: %s, Net Income: %s\n", r.TotalIncome.String(), r.TotalExpense.String(), r.NetIncome.String()))
	str.WriteString("Total Income Per Person (%% of total income)\n")
	sortedNameKeys := sortKeys(maps.Keys(r.TotalIncomePerPayer))
	for _, name := range sortedNameKeys {
		str.WriteString(fmt.Sprintf("%s: %s (%.2f%%)\n", name, r.TotalIncomePerPayer[name], 100*float64(r.TotalIncomePerPayer[name].Cents())/float64(r.TotalIncome.Cents())))
	}
	str.WriteString("Total Expense Per Person (%% of total expenses)\n")
	for _, name := range sortedNameKeys {
		str.WriteString(fmt.Sprintf("%s: %s (%.2f%%)\n", name, r.TotalExpensePerPayer[name], 100*float64(r.TotalExpensePerPayer[name].Cents())/float64(r.TotalExpense.Cents())))
	}
	str.WriteString("Total Expense Per Category (%% of total expenses)\n")
	expensePerDescription := r.CalculateTotalExpensePerDescription()
	sortedKeys := sortKeys(maps.Keys(expensePerDescription))
	for _, key := range sortedKeys {
		str.WriteString(fmt.Sprintf("%s: %s (%.2f%%)\n", key, expensePerDescription[key].String(), 100*float64(expensePerDescription[key].Cents())/float64(r.TotalExpense.Cents())))
	}
	str.WriteString("Time,Amount,Description,Name\n")
	for _, income := range r.SortIncomes() {
		str.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", income.Time, income.Amount.String(), income.Description, income.PaidBy))
	}
	for _, expense := range r.SortExpenses() {
		str.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", expense.Time, expense.Amount.String(), expense.Description, expense.PaidBy))
	}
	return str.String()
}

// WriteCSV writes the report to a CSV
func (r MultiPayerReport) WriteCSV(writer io.Writer) error {
	fmt.Fprint(writer, "Time,Amount,Description,Name")
	for _, income := range r.SortIncomes() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s,%s", income.Time, float64(income.Amount.Cents())/100, income.Description, income.PaidBy)
	}
	for _, expense := range r.SortExpenses() {
		fmt.Fprintf(writer, "\n%s,%.2f,%s,%s", expense.Time, float64(expense.Amount.Cents())/100, expense.Description, expense.PaidBy)
	}
	return nil
}

// SortIncomes sort the incomes in the report from largest to smallest
func (r MultiPayerReport) SortIncomes() []transaction.PayerTransaction {
	var incomes []transaction.PayerTransaction
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) > 0 {
			incomes = append(incomes, transaction)
		}
	}
	return sortPayerTransactions(incomes)
}

// SortExpenses sorts expenses in the report from largest expense to smallest expense
func (r MultiPayerReport) SortExpenses() []transaction.PayerTransaction {
	var expenses []transaction.PayerTransaction
	for _, transaction := range r.transactions {
		if transaction.Amount.Cmp(currency.NewEuro(0.0)) < 0 {
			expenses = append(expenses, transaction)
		}
	}
	return sortPayerTransactions(expenses)
}

func sortPayerTransactions(transactions []transaction.PayerTransaction) []transaction.PayerTransaction {
	slices.SortFunc(transactions, func(a, b transaction.PayerTransaction) int {
		if math.Abs(float64(a.Amount.Cents())) <= math.Abs(float64(b.Amount.Cents())) {
			return 1
		} else {
			return -1
		}
	})
	return transactions
}

// Transactions returns a copy of the transactions from the report
func (r MultiPayerReport) Transactions() []transaction.PayerTransaction {
	var transactions []transaction.PayerTransaction
	for _, tx := range r.transactions {
		transactions = append(transactions, transaction.PayerTransaction{Time: tx.Time, Amount: tx.Amount, Description: tx.Description, PaidBy: tx.PaidBy})
	}
	return transactions
}
