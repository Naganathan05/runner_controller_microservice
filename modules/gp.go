package modules

import (
	"encoding/json"
	"evolve/util"
	"fmt"
	"strings"
)

type GP struct {
	Algorithm          string    `json:"algorithm"`
	Arity              int       `json:"arity"`
	Operators          []string  `json:"operators"`
	ArgNames           []string  `json:"argNames"`
	IndividualType     string    `json:"individualType"`
	Expr               string    `json:"expr"`
	RealFunction       string    `json:"realFunction"`
	Min                int       `json:"min"`
	Max                int       `json:"max"`
	IndividualFunction string    `json:"individualFunction"`
	PopulationFunction string    `json:"populationFunction"`
	SelectionFunction  string    `json:"selectionFunction"`
	TournamentSize     int       `json:"tournamentSize"`
	ExprMut            string    `json:"expr_mut"`
	CrossoverFunction  string    `json:"crossoverFunction"`
	TerminalProb       float64   `json:"terminalProb"`
	MutationFunction   string    `json:"mutationFunction"`
	MutationMode       string    `json:"mutationMode"`
	MateHeight         int       `json:"mateHeight"`
	MutHeight          int       `json:"mutHeight"`
	Weights            []float64 `json:"weights"`
	PopulationSize     int       `json:"populationSize"`
	Generations        int       `json:"generations"`
	Cxpb               float64   `json:"cxpb"`
	Mutpb              float64   `json:"mutpb"`
	Mu                 int       `json:"mu"`
	Lambda             int       `json:"lambda"`
	IndividualSize     int       `json:"individualSize"`
	HofSize            int       `json:"hofSize"`
	ExprMutMin         int       `json:"expr_mut_min"`
	ExprMutMax         int       `json:"expr_mut_max"`
}

func GPFromJSON(jsonData map[string]any) (*GP, error) {
	gp := &GP{}
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonDataBytes, gp); err != nil {
		return nil, err
	}
	return gp, nil
}

func (gp *GP) validate() error {
	if err := util.ValidateAlgorithmName(gp.Algorithm); err != nil {
		return err
	}
	// TODO: Validate remaining fields.
	return nil
}

func (gp *GP) imports() string {
	return strings.Join([]string{
		"import operator",
		"import math",
		"import random",
		"import numpy",
		"import matplotlib.pyplot as plt",
		"import networkx as nx",
		"from functools import partial",
		"from deap import algorithms, base, creator, tools, gp, cma",
		"from scoop import futures",
	}, "\n")
}

func (gp *GP) evalFunction() string {
	var evalFunc string
	evalFunc += "def evalSymbReg(individual, points, realFunction):\n"
	evalFunc += "\t# Transform the tree expression in a callable function\n"
	evalFunc += "\tfunc = toolbox.compile(expr=individual)\n"
	evalFunc += "\t# Evaluate the mean squared error between the expression\n"
	evalFunc += "\t# and the real function : x**4 + x**3 + x**2 + x\n"
	evalFunc += "\tsqerrors= ((func(x) - eval(realFunction))**2 for x in points)\n"
	evalFunc += "\treturn (math.fsum(sqerrors) / len(points),)\n\n"
	return evalFunc
}

func (gp *GP) addPrimitivesToPSET() string {
	primitiveSource := map[string]map[string]any{
		"add": {"arity": 2, "code": "operator.add"},
		"sub": {"arity": 2, "code": "operator.sub"},
		"mul": {"arity": 2, "code": "operator.mul"},
		"div": {"arity": 2, "code": "protectedDiv"},
		"neg": {"arity": 1, "code": "operator.neg"},
		"cos": {"arity": 1, "code": "math.cos"},
		"sin": {"arity": 1, "code": "math.sin"},
		"lf":  {"arity": 1, "code": "lf"},
	}

	var primitives string
	for _, operator := range gp.Operators {
		primitive := primitiveSource[operator]
		primitives += fmt.Sprintf("pset.addPrimitive(%s, %d)\n", primitive["code"], primitive["arity"])
	}
	return primitives
}

func (gp *GP) Code() (string, error) {
	if err := gp.validate(); err != nil {
		return "", err
	}

	var code string
	code += gp.imports() + "\n\n"
	code += gp.evalFunction() + "\n\n"

	code += "toolbox = base.Toolbox()\n"
	code += "pset = gp.PrimitiveSet('MAIN', 1)\n\n"

	code += "def protectedDiv(left, right):\n"
	code += "\ttry:\n"
	code += "\t\treturn left / right\n"
	code += "\texcept ZeroDivisionError:\n"
	code += "\t\treturn 1\n\n"

	code += "def lf(x):\n"
	code += "\treturn 1 / (1 + numpy.exp(-x))\n\n"

	code += gp.addPrimitivesToPSET() + "\n\n"

	return code, nil
}
