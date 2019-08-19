package main

import (
	"fmt"
	"bufio"
	"log"
	"os"
	_ "io"
)

// some kind of dictionary to keep account of our variables
type variableTree struct {
	Name string
	Rules []string
}

var (
	variables []variableTree
	state = "EMPTY"
	cc string
	outcomes []string
	superOutcomes []string
)

func check(e error){
	if e != nil {
		log.Fatal(e)
	}
}

func varParser(line string) {
	var varName string
	// get the name of the variable!!
	for index, char := range line{
		if char == '%' && index != 0 {
			break
		} else if char == '%' {
			continue
		} else {
			varName += string(char)
		}
	}
	// if we already were in the READV state check for the
	// last variableTree added in the forest
	// if it's this one change the state and goto next line
	// otherwise err
	if state == "READV" {
		if variables[len(variables)-1].Name == varName{
			state = "EMPTY"
		} else {
			log.Fatal("Cannot have nested variable declarations!!")
		}
	} else {
		// create a variableTree in the variableForest for the new var
		var rules []string
		newVar := variableTree{varName, rules}
		variables = append(variables, newVar)
		// modify the state to READ_VAR and point to the place where the rules should be saved
		state="READV"
	}
	//	fmt.Printf("variable %s begin/end\n", varName)
}

func rulesParser(line string) {
	// as pc doesn't allow variables nested declarations
	// the last variable added to the forest is always going to be the one
	// that's read at a specific moment
	variables[len(variables)-1].Rules = append(variables[len(variables)-1].Rules, line)
}

func findVarByName(name string) int {
	for index, variable := range variables {
		if variable.Name == name {
			return index
		}
	}
	return -1
}

func evalRule(rule string){
	// rn localOutcome only has the literal of this rule
	// but when we will implement deriving rules it's gonna have
	// more outcomes into it
	// add rules outcomes to the local outcomes
	var localOutcome []string
	localOutcome = append(localOutcome, rule)
	// add local outcomes to super outcomes
	for _, out := range localOutcome {
		superOutcomes = append(superOutcomes, out)
	}
}

func apply(v1, v2 []string) []string{
	var ret []string
	if v1 == nil {
		v1 = append(v1, "")
	} else if v2 == nil {
		v2 = append(v2, "")
	}
	for _, val1 := range v1{
		for _, val2 := range v2{
			ret = append(ret, val1+val2)
		}
	}
	return ret
}

func eval(statement string) {
	pointer := 0
	for pointer < len(statement) {
		cc = string(statement[pointer])
		if cc == "%" {
			varName := ""
			pointer += 1				
			cc = string(statement[pointer])
			for cc != "%" {
				varName += cc
				pointer += 1
				cc = string(statement[pointer])
			}
			//fmt.Printf("~%s~", varName)
			// find the variable's index
			index := findVarByName(varName)
			if index < 0 {
				log.Fatal("Variable ", varName, " not declared!")
			}
			//fmt.Println(index)
			// now evaluate each and every rule defined for this
			// variable
			for _, rule := range variables[index].Rules {
				evalRule(rule)
			}
			/*
			fmt.Printf("Outcomes array: ")
			for _,out := range outcomes{
				fmt.Printf("%s, ", out)
			}
			fmt.Printf("\nsuperOutomes array: ")
			for _,out := range superOutcomes{
				fmt.Printf("%s, ", out)
			}*/
			// appky superoutcomes to outcomes
			outcomes = apply(outcomes, superOutcomes)
			superOutcomes = nil
			/*fmt.Printf("\nAfter apply outcomes looks like this: ")
			for _,out := range outcomes{
				fmt.Printf("%s, ", out)
			}
			fmt.Printf("\n")*/
		} else {
			// this is for literal expressions
			//			fmt.Printf(cc)
			// we add to the result slice the outcome
			if outcomes == nil {
				outcomes = append(outcomes, cc)
			} else {
				for i,_ := range outcomes {
					outcomes[i] += cc
				}
			}
		}
		pointer += 1
	}
	//fmt.Printf("\n")
	//	fmt.Println(statement)
}

func parser(line string) {
	//fmt.Println(line)
	if line == "" {
		//fmt.Println("Empty line")
	} else {
		switch line[0] {
		case '>':
			toImport := line[1:]
			imported, err := os.Open(toImport + ".pc")
			check(err)
			defer imported.Close()
			importScan := bufio.NewScanner(imported)
			for importScan.Scan() {
				parser(importScan.Text())
			}
		case ';':
			// we don't do nothing to the comments (rn)
			//fmt.Println("comment")
		case '%':
			varParser(line)
		case '{':
			eval(line[1:(len(line)-1)])
			fmt.Println("After the eval of the statement: ")
			for _, outcome := range outcomes {
				fmt.Println(outcome)
			}
			outcomes = nil
			fmt.Println()
			//fmt.Println("A thing to eval")
		default:
			if state == "READV" {
				rulesParser(line)
				//fmt.Println("^this is a rule")
			} else {
				//fmt.Println("Not interesting")
			}
		}
	}
}

func main() {
	// open the file to read
	// check first if we are given a file to read
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./passwordCrack <file>\nYou must give a file")
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// open the file to write
	dstFile, err := os.Create("parsed.txt")
	check(err)
	writer := bufio.NewWriter(dstFile)
	_ = writer

	// scan each line and feed it to our parser function
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parser(scanner.Text())
	}

	
	// check to see if all the variables were added -DEBUG-
	/*	fmt.Println()
	for _, variable := range variables {
		fmt.Println(variable.Name)
		for _, rule := range variable.Rules {
			fmt.Println(rule)
		}
	}*/
}
