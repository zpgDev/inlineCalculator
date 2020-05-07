/**
	Inline calculator
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Splited struct {
	operator string
	nums []string
	parsedNums []float64
}

// The sequence of operators matters!
var operators = [6]string{"+", "-", "*", "/", ":", "^"}

const headInfo = `Inline calculator
Usage modes:
	Console: icalc <operand1><operator><operand2>[<operator><operandN>...] | <command>
	Interactive: <operand1><operator><operand2>[<operator><operandN>...] | <command>`

var helpInfo = headInfo + `

For interactive mode run icalc without arguments.

Commands:
	-h, --help		for more information about a commands
	-o, --operators		list of supported operators
	q, quit, exit		exit interactive mode
`

var operatorsInfo = headInfo + `

Supported operators:
	+	addition
	-	subtraction
	*	multiplication
	/, :	division
	^	exponentiation
`

// convert args to float64
func parseOperands(c []string) ([]float64, error) {
	var result []float64
	for _, num := range c {
		parsedNum, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return result, err
		}
		result = append(result, parsedNum)
	}
	return result, nil
}

// split expression by operator
func splitParams(param string) (Splited, error) {
	var returned Splited
	// remove all spaces in expression
	param = strings.Replace(param, " ", "", -1)
	// bringing to a single operator
	param = strings.Replace(param, ":", "/", -1)
	if param == "" {
		return returned, errors.New("error: no params found")
	}

	for _, operator := range operators {
		param = strings.Replace(param, "." + operator, operator, -1)
		pattern := `\d+\` + operator
		match, err := regexp.MatchString(pattern, param)
		if err != nil {
			return returned, err
		}
		if match {
			nums := strings.Split(param, operator)
			if nums[len(nums)-1] == "" {
				nums = nums[:len(nums)-1]
			}
			if operator == "-" {
				for index, sp := range nums {
					if sp == "" {
						nums[index + 1] = "-" + nums[index + 1]
						nums = remove(nums, index)
					}
				}
			}
			splited := Splited{
				operator: operator,
				nums:    nums,
			}
			return splited, nil
		}
	}

	return returned, errors.New("error: no operator found")
}

// check input commands
func checkCommands(command string) string {
	res := ""
	switch command {
		case "exit", "quit", "q":
			os.Exit(0)
		case "-h", "--help":
			res = helpInfo
		case "-o", "--operators":
			res = operatorsInfo
	}
	return res
}

func parseParams(params string) (float64, error) {
	result := 0.0
	err := error(nil)

	splited, err := splitParams(params)
	if err != nil {
		return result, err
	}

	if len(splited.nums) < 2 {
		return result, errors.New("error: not enough arguments")
	}

	splited.parsedNums = make([]float64, len(splited.nums))

	for index, val := range splited.nums {
		// checking for expressions in operands
		match, err := regexp.MatchString(`\d+[+\-*/:^]+\d+`, val)
		if err != nil {
			return result, err
		}
		if match {
			matchedCalc, err := parseParams(val)
			if err != nil {
				return result, err
			}
			// set result of expression for next calculation
			splited.parsedNums[index] = matchedCalc
		} else {
			parsedNum, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return result, err
			}
			splited.parsedNums[index] = parsedNum
		}
	}
	result, err = calculate(splited)
	if err != nil {
		return result, err
	}

	return result, err
}

func calculate(splited Splited) (float64, error) {
	result := 0.0
	var nums []float64
	err := error(nil)
	if len(splited.parsedNums) > 1 {
		nums = splited.parsedNums
	} else {
		nums, err = parseOperands(splited.nums)
		if err != nil {
			return 0.0, err
		}
	}

	switch splited.operator {
		case "*":
			result = multiply(nums)
		case "/":
			result, err = divide(nums)
		case "^":
			result = pow(nums)
		case "+":
			result = add(nums)
		case "-":
			result = subtract(nums)
	}

	if err != nil {
		return 0.0, err
	}

	return result, nil
}

func multiply(nums []float64) float64 {
	result := 1.0
	for _, num := range nums {
		result *= num
	}
	return result
}

func divide(nums []float64) (float64, error) {
	var result float64
	for index, num := range nums {
		if index > 0 && num == 0.0 {
			return 0.0, errors.New("error: you tried to divide by zero")
		}
		if result == 0 {
			result = num
		} else {
			result /= num
		}
	}
	return result, nil
}

func pow(nums []float64) float64 {
	var result float64
	for index, num := range nums {
		if index == 0 && result == 0 {
			result = num
		} else {
			result = math.Pow(result, num)
		}
	}
	return result
}

func add(nums []float64) float64 {
	result := 0.0
	for _, num := range nums {
		result += num
	}
	return result
}

func subtract(nums []float64) float64 {
	var result float64
	for index, num := range nums {
		if index == 0 && result == 0 {
			result = num
		} else {
			result -= num
		}
	}
	return result
}

// remove element from slice
func remove(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func process(params string, interactive bool) {
	command := checkCommands(params)
	if command != "" {
		fmt.Println(command)
	} else {
		res, err := parseParams(params)
		if err != nil {
			fmt.Println(err)
		} else {
			if interactive {
				fmt.Println("=", res)
			} else {
				fmt.Println(res)
			}
		}
	}
}

func main() {
	args := os.Args
	if len(args) > 1 {
		// console mode
		process(args[1], false)
		os.Exit(0)
	}

	// interactive mode
	fmt.Println(headInfo)
	fmt.Println("Type --help for more info")
	for {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("icalc> ")
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			process(text, true)
			fmt.Print("icalc> ")
		}
	}
}