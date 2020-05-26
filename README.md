### Inline calculator on GoLang

This is free software with ABSOLUTELY NO WARRANTY.

The inline calculator can work in two modes, console and interactive:

**Console mode:**
Run inline calculator with expression or command

``go run icalc.go <operand1><operator><operand2>[<operator><operandN>...] | <command>``

Example:
``go run icalc.go 2+2*2``

**Interactive mode:**
Run inline calculator without arguments

``go run icalc.go``

``icalc> <operand1><operator><operand2>[<operator><operandN>...] | <command>``

**Available commands:**

````
-h, --help		for more information about a commands
-o, --operators		list of supported operators
h, history		history of calculations in interactive mode
c, cls, clear		clear terminal in interactive mode
q, quit, exit		exit interactive mode
````

**Supported operators:**

````
+	addition
-	subtraction
*	multiplication
/, :	division
^	exponentiation
````