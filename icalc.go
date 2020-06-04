/**
	Inline calculator
	This is free software with ABSOLUTELY NO WARRANTY.
	Author: Pavlo Zubkov (zubkov.dev@gmail.com)
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
var operators = [7]string{"+", "-", "%", "*", "/", ":", "^"}
var operatorsPattern = `+\-%*/:^`

var isParentheses bool
// max available iterations in recursive calls
var iteration = 1000

const termText = " icalc> "

const headInfo = `Inline calculator
(c) 2020 Pavlo Zubkov
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
	%	modulo
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

// parsing and calculation expressions in parentheses
func parseParentheses(params string) (string, error) {
	var err error

	// check iteration limit
	if iteration <= 0 {
		iteration = 1000
		return "", setError("iteration limit is reached")
	}
	iteration--

	// TODO: optimize
	outerParentheses := regexp.MustCompile(`^\(([\d.` + operatorsPattern + `]+|(?:[\d.` + operatorsPattern + `]*\([\d.` + operatorsPattern + `()]*\)[\d.` + operatorsPattern + `]*)+)\)$`)

	// remove parent parentheses (5+5) if exists
	oRes := outerParentheses.FindStringSubmatch(params)
	if len(oRes) == 2 {
		params = oRes[1]

		// check if another parent parentheses present
		oRes = outerParentheses.FindStringSubmatch(params)
		if len(oRes) == 2 {
			return parseParentheses(params)
		}
	}

	// calculate parentheses expression and replace it by result
	innerParentheses := regexp.MustCompile(`\(([\d.` + operatorsPattern + `]+)\)`)
	params = innerParentheses.ReplaceAllStringFunc(params, func(s string) string {
		var br float64
		var numberOnly bool
		isParentheses = false

		// remove parentheses from expression
		sRes := innerParentheses.FindStringSubmatch(s)
		if len(sRes) == 2 {
			s = sRes[1]
		}

		// if only number in parentheses
		numberOnly, err = regexp.MatchString(`^([\d.]+)$`, s)
		if err != nil {
			return ""
		}
		if numberOnly {
			return s
		}

		// parse and calculate expression
		br, err = parseParams(s)
		if err != nil {
			return ""
		}

		return fmt.Sprint(br)
	})

	// check if parentheses present
	isParentheses, err = regexp.MatchString(`\(.*\)`, params)
	if err != nil {
		return params, err
	}
	if isParentheses {
		return parseParentheses(params)
	}

	return params, err
}

func parseParams(params string) (float64, error) {
	var result float64
	var err error
	var splited Splited

	// check iteration limit
	if iteration <= 0 {
		iteration = 1000
		return result, setError("iteration limit is reached")
	}
	iteration--

	if isParentheses {
		params, err = parseParentheses(params)
		if err != nil {
			return result, err
		}
	}

	splited, err = splitParams(params)
	if err != nil {
		return result, err
	}

	if len(splited.nums) < 2 {
		return result, setError("not enough arguments")
	}

	splited.parsedNums = make([]float64, len(splited.nums))

	for index, val := range splited.nums {
		// checking for expressions in operands
		match, err := regexp.MatchString(`\d+(\.*)[` + operatorsPattern + `]+(\.*)\d+`, val)

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
	var result float64
	var nums []float64
	var err error

	if len(splited.parsedNums) > 1 {
		nums = splited.parsedNums
	} else {
		nums, err = parseOperands(splited.nums)
	}

	if err == nil {
		switch splited.operator {
			case "*":
				result = multiply(nums)
			case "/":
				result, err = divide(nums)
			case "^":
				result = pow(nums)
			case "%":
				result, err = modd(nums)
			case "+":
				result = add(nums)
			case "-":
				result = subtract(nums)
			default:
				err = setError("unsupported operator")
		}
	}

	return result, err
}

// math functions start
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

func modd(nums []float64) (float64, error) {
	var result float64
	for index, num := range nums {
		if index > 0 && num == 0.0 {
			return 0.0, setError("Modulo by zero")
		}
		if index == 0 && result == 0 {
			result = num
		} else {
			result = math.Mod(result, num)
		}
	}
	return result, nil
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
// math functions end

// remove element from slice
func remove(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

// check input commands in bash mode
func checkCommands(command string) string {
	res := ""
	switch command {
	case "-h", "--help":
		res = helpInfo
	case "-o", "--operators":
		res = operatorsInfo
	default:
		res = "Command not found"
	}
	return res
}

// check input commands in interactive
func checkInteractiveCommands(command string, term *terminal.Terminal) string {
	res := ""
	switch command {
	case "exit", "quit", "q":
		exitCommand(term)
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
	default:
		res = "\nCommand not found"
	}
	return res
}

// exit interactive mode
func exitCommand(term *terminal.Terminal) {
	_, _ = term.Write([]byte("Exit\r\n"))
	term.ReleaseFromStdInOut()
	fmt.Println("")
	os.Exit(0)
}

// check if input is command
func checkIsCommand(params string) (bool, error) {
	return regexp.MatchString(`^([-a-zA-Z]+)+([^\d])*$`, params)
}

// bash mode
func process(params string) {
	res := 0.0
	var err error
	var isCommand bool
	command := ""

	// check if command
	isCommand, err = checkIsCommand(params)

	if err == nil {
		if isCommand {
			command = checkCommands(params)
			if command != "" {
				fmt.Println(command)
			}
		} else {
			var operandOnly bool
			operandOnly, err = checkInput(params)
			if err == nil {
				params, err = cleanParams(params)
				if err == nil {
					if operandOnly {
						// if number only
						res, err = strconv.ParseFloat(params, 64)
						if err != nil {
							err = parseError(err)
						}
					} else {
						res, err = parseParams(params)
					}
					if err == nil {
						fmt.Println(res)
					}
				}
			}
		}
	}

	if err != nil {
		fmt.Println(setFgColor(RED, setBoldError(err)))
	}
}

func interactiveProcess(params string, term *terminal.Terminal) {
	res := 0.0
	var err error
	var isCommand bool
	command := ""

	// check if command
	isCommand, err = checkIsCommand(params)

	if err == nil {
	 	if isCommand {
			command = checkInteractiveCommands(params, term)
			if command != "" {
				if command != "-clear-" {
					fmt.Println(command)
				}
			}
		} else {
			var operandOnly bool
			operandOnly, err = checkInput(params)
			if err == nil {
				params, err = cleanParams(params)
				if err == nil {
					if operandOnly {
						// if only number
						res, err = strconv.ParseFloat(params, 64)
						if err != nil {
							err = parseError(err)
						}
					} else {
						res, err = parseParams(params)
					}
					if err == nil {
						fmt.Println(setBold("="), setBoldFloat(res))
					}
				}
			}
		}
	}

	if err != nil {
		fmt.Println(setFgColor(RED, setBoldError(err)))
		res = math.NaN()
	}

	if command != "-clear-" {
		fmt.Println("")
	}

	isParentheses = false

	// add result to history
	term.AddResultHistory(res)
}

// clean input params
func cleanParams(params string) (string, error) {
	var err error

	// remove all spaces in expression
	params = strings.Replace(params, " ", "", -1)

	if isParentheses {
		params, err = cleanParentheses(params)
	}

	if params == "" {
		return params, setError("no params found")
	}

	// bringing to a single operator
	params = strings.Replace(params, ":", "/", -1)

	return params, err
}

// clean input params
func cleanParentheses(params string) (string, error) {
	var err error

	// check iteration limit
	if iteration <= 0 {
		iteration = 1000
		return  "", setError("iteration limit is reached")
	}
	iteration--

	prPattern := regexp.MustCompile(`(?:\()([\s\d.]*)(?:\))`)
	params = prPattern.ReplaceAllStringFunc(params, func(s string) string {
		p := prPattern.FindStringSubmatch(s)
		if len(p) == 2 {
			s = p[1]
		}

		return s
	})

	if prPattern.MatchString(params) {
		return cleanParentheses(params)
	}

	return params, err
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

// clear terminal
func clear() {
	fmt.Print("\033[H\033[2J")
}

// check for correct input
func checkInput(params string) (bool, error) {
	var err error
	var wrongSpaces bool
	var containSymbol bool
	var invalidSyntax bool

	// check parentheses
	openPattern := regexp.MustCompile(`\(`)
	closePattern := regexp.MustCompile(`\)`)
	openB := openPattern.FindAllString(params, -1)
	closeB := closePattern.FindAllString(params, -1)
	if len(openB) != len(closeB) {
		return false, setError("Invalid syntax: Parentheses mismatch")
	} else if len(openB) > 0 && len(closeB) > 0 {
		isParentheses = true
	}

	// check if expression contains symbols
	containSymbol, err = regexp.MatchString(`([a-zA-Z]+)`, params)
	if err != nil {
		return false, err
	}
	if containSymbol {
		err = setError("Invalid syntax: The expression contains symbols")
		return false, err
	}

	// check invalid syntax
	invalidSyntax, err = regexp.MatchString(`\)[\s\d.]*\(|(\s*[\d.]+\s*)\(|\)(\s*[\d.]+\s*)`, params)
	if err != nil {
		return false, err
	}
	if invalidSyntax {
		err = setError("Invalid syntax")
		return false, err
	}

	// check wrong spaces
	wrongSpaces, err = regexp.MatchString(`(?:\d|\.)+(\s)+(?:\d|\.)+`, params)
	if err != nil {
		return false, err
	}
	if wrongSpaces {
		return false, setError("Invalid syntax: Wrong space entered")
	}

	// if number only
	operandOnly, err := regexp.MatchString(`^(\()*(\s)*(\d|\.)+(\.)*(\d)*(\s)*(\))*$`, params)

	return operandOnly, err
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

	term, termErr := terminal.NewWithStdInOut()
	if termErr != nil {
		panic(termErr)
	}
	defer term.ReleaseFromStdInOut() // defer this
	fmt.Println("")

	termFText := setBgColor(CYAN, setFgColor(YELLOW, setBold(termText))) + " "
	term.SetPrompt(termFText)

	for {
		line, err := term.ReadLine()
		if err != nil {
			if strings.Contains(err.Error(), "control-c break") {
				exitCommand(term)
			} else {
				fmt.Println("error: ", err)
				break
			}
		}
		text := strings.TrimSpace(line)
		interactiveProcess(text, term)
	}
}