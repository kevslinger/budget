package report

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kevslinger/budget/currency"
	"github.com/kevslinger/budget/transaction"
)

type Report interface {
	fmt.Stringer
	Save(filename string) error
	WriteCSV(writer io.Writer) error
}

// CombineReports merges a slice of reports into a single report
// It assumes that all reports are of the same type
func CombineReports(reportName string, reports []Report) (Report, error) {
	switch reports[0].(type) {
	case BasicReport:
		return CombineBasicReports(reportName, reports)
	case MultiPayerReport:
		return CombineMultiPayerReports(reportName, reports)
	}
	return BasicReport{}, fmt.Errorf("unknown report type: %T", reports[0])
}

// ReadBudgetReportFromFile reads in a CSV file with transactions, and parses them to create a report
func ReadBudgetReportFromFile(reportName string, path string) (Report, error) {
	file, err := os.Open(path)
	if err != nil {
		return BasicReport{}, fmt.Errorf("error opening budget report file: %w", err)
	}
	reader := csv.NewReader(file)
	line, err := reader.Read()
	if err != nil {
		return BasicReport{}, fmt.Errorf("error reading budet report file: %w", err)
	}
	file.Close()
	switch len(line) {
	case 3:
		return ReadDefaultBudgetReportFromFile(reportName, path)
	case 4:
		return ReadMultiPayerBudgetReportFromFile(reportName, path)
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

func sortKeys(m iter.Seq[string]) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
