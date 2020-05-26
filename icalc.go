/**
	Inline calculator
	This is free software with ABSOLUTELY NO WARRANTY.
	Author: Pavlo Zubkov
	(c) 2020
 */

package main

import (
	"./terminal"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	BLACK = iota
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

type Splited struct {
	operator string
	nums []string
	parsedNums []float64
}

// The sequence of operators matters!
var operators = [6]string{"+", "-", "*", "/", ":", "^"}

const termText = " icalc> "

const headInfo = `Inline calculator
This is free software with ABSOLUTELY NO WARRANTY.
Usage:
  icalc <operand1><operator><operand2>[<operator><operandN>...] | <command>
`

var helpInfo = headInfo + `

For interactive mode run icalc without arguments.

Commands:
	-h, --help		for more information about a commands
	-o, --operators		list of supported operators
	h, history		history of calculations in interactive mode
	c, cls, clear		clear terminal in interactive mode
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
			err = parseError(err)
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
		return returned, setError("no params found")
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

	return returned, setError("no operator found")
}

// check input commands in bash mode
func checkCommands(command string) string {
	res := ""
	switch command {
		case "-h", "--help":
			res = "\n" + helpInfo
		case "-o", "--operators":
			res = "\n" + operatorsInfo
	}
	return res
}

// check input commands in interactive
func checkInteractiveCommands(command string, term *terminal.Terminal) string {
	res := ""
	switch command {
		case "exit", "quit", "q":
			_, _ = term.Write([]byte("Exit\r\n"))
			term.ReleaseFromStdInOut()
			os.Exit(0)
		case "clear", "cls", "c":
			res = "-clear-"
			clear()
		case "history", "h":
			hist := term.GetHistory()
			fmt.Println("Calculations history:")
			if len(hist) > 0 {
				for _, h := range hist {
					res += "\n" + h
				}
			} else {
				res = "\nNo history found"
			}
		case "-h", "--help":
			res = "\n" + helpInfo
		case "-o", "--operators":
			res = "\n" + operatorsInfo
	}
	return res
}

func parseParams(params string) (float64, error) {
	result := 0.0
	err := error(nil)

	// check wrong spaces
	wrongSpaces, _ := regexp.MatchString(`(\d|\.)+(\s+)(\d|\.)+`, params)
	if wrongSpaces {
		return result, setError("invalid syntax")
	}

	// if number only
	operandOnly, _ := regexp.MatchString(`^(\d|\.)+(\.)*(\d)*$`, params)
	if operandOnly {
		f, err := strconv.ParseFloat(params, 64)
		if err != nil {
			return math.NaN(), err
		}
		return f, nil
	}

	splited, err := splitParams(params)
	if err != nil {
		return result, err
	}

	if len(splited.nums) < 2 {
		return result, setError("not enough arguments")
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
				err = parseError(err)
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
			return 0.0, setError("you tried to divide by zero")
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

// bash mode
func process(params string) {
	res := 0.0
	var err error
	command := checkCommands(params)
	if command != "" {
		fmt.Println(command)
	} else {
		res, err = parseParams(params)
		if err != nil {
			fmt.Println(setFgColor(RED, setBoldError(err)))
		} else {
			fmt.Println(res)
		}
	}
}

func interactiveProcess(params string, term *terminal.Terminal) {
	res := 0.0
	var err error
	command := checkInteractiveCommands(params, term)
	if command != "" {
		if command != "-clear-" {
			fmt.Println(command)
		}
	} else {
		res, err = parseParams(params)
		if err != nil {
			fmt.Println(setFgColor(RED, setBoldError(err)))
		} else {
			fmt.Println(setBold("="), setBoldFloat(res))
		}
	}
	if command != "-clear-" {
		fmt.Println("")
	}

	if err != nil {
		res = math.NaN()
	}
	term.AddResultHistory(res)
}

func parseError(err error) error {
	errString := err.Error()
	errSlice := strings.Split(errString, ": ")
	if len(errSlice) > 2 {
		err = setError(errSlice[1] + " - " + errSlice[2])
	}
	return err
}

func setError(text string) error {
	return errors.New("error: " + text)
}

func setFgColor(color int, text string) string {
	return fmt.Sprintf("%s%s\033[0m", fmt.Sprintf("\033[3%dm", color), text)
}

func setBgColor(color int, text string) string {
	return fmt.Sprintf("%s%s\033[0m", fmt.Sprintf("\033[4%dm", color), text)
}

func setBold(text string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}

func setBoldError(err error) string {
	return fmt.Sprintf("\033[1m%s\033[0m", err)
}

func setBoldFloat(res float64) string {
	return fmt.Sprintf("\033[1m%v\033[0m", res)
}

func clear() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	args := os.Args
	if len(args) > 1 {
		// bash mode
		process(args[1])
		os.Exit(0)
	}

	// interactive mode
	clear()
	fmt.Println(headInfo)
	fmt.Println("Type --help for more info")
	termFText := setBgColor(CYAN, setFgColor(YELLOW, setBold(termText))) + " "

	term, termErr := terminal.NewWithStdInOut()
	if termErr != nil {
		panic(termErr)
	}
	defer term.ReleaseFromStdInOut() // defer this
	fmt.Println("")
	term.SetPrompt(termFText)

	for {
		line, err := term.ReadLine()
		if err != nil {
			if strings.Contains(err.Error(), "control-c break") {
				_, _ = term.Write([]byte(line + "\r\n"))
				continue
			} else {
				fmt.Println("error: ", err)
				break
			}
		}
		text := strings.TrimSpace(line)
		interactiveProcess(text, term)
	}
}