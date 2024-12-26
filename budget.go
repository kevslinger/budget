package budget

import (
	"fmt"
)

func Main() int {
	fmt.Printf("Welcome to the budget tracker app!\nEnter your income from the last month (without currency symbol): ")
	var income float64
	_, err := fmt.Scanln(&income)
	if err != nil {
		fmt.Println("There was an error recording your income! Please restart the program and input a valid number. Error: ", err)
		return 1
	}
	fmt.Printf("Congrats on earning %.2f last month!\n", income)
	fmt.Printf("Please also enter your total expenses last month (in the same currency, without currency symbol): ")
	var expenses float64
	_, err = fmt.Scanln(&expenses)
	if err != nil {
		fmt.Println("There was an error recording your expenses. Please restart the program and input a valid number for expenses. Error: ", err)
		return 1
	}

	fmt.Printf("After paying your expenses for the month, you were left with %.2f!\n", income-expenses)
	return 0
}
