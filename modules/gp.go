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
	Min                int       `json:"min_"`
	Max                int       `json:"max_"`
	IndividualFunction string    `json:"individualFunction"`
	PopulationFunction string    `json:"populationFunction"`
	SelectionFunction  string    `json:"selectionFunction"`
	TournamentSize     int       `json:"tournamentSize"`
	ExprMut            string    `json:"expr_mut"`
	CrossoverFunction  string    `json:"crossoverFunction"`
	TerminalProb       float64   `json:"terminalProb,omitempty"`
	MutationFunction   string    `json:"mutationFunction"`
	MutationMode       string    `json:"mutationMode,omitempty"`
	MateHeight         int       `json:"mateHeight"`
	MutHeight          int       `json:"mutHeight"`
	Weights            []float64 `json:"weights"`
	PopulationSize     int       `json:"populationSize"`
	Generations        int       `json:"generations"`
	Cxpb               float64   `json:"cxpb"`
	Mutpb              float64   `json:"mutpb"`
	Mu                 int       `json:"mu,omitempty"`
	Lambda             int       `json:"lambda_,omitempty"`
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
		"import os",
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

func (gp *GP) renameArgs() string {
	argDict := make(map[string]string)
	for i, name := range gp.ArgNames {
		argDict[fmt.Sprintf("ARG%d", i)] = name
	}

	var code string

	var argString string
	for k, v := range argDict {
		argString += fmt.Sprintf("'%s': '%s', ", k, v)
	}
	// Remove trailing comma and space
	if len(argString) > 2 {
		argString = argString[:len(argString)-2]
	}

	code += fmt.Sprintf("arg_dict = {%s}\n", argString)
	code += "pset.renameArguments(**arg_dict)\n\n"
	return code
}

func (gp *GP) selectionFunction() string {
	// TODO: Add support for other selection functions.
	switch gp.SelectionFunction {
	case "selTournament":
		return fmt.Sprintf("toolbox.register('select', tools.%s, tournsize=%d)\n", gp.SelectionFunction, gp.TournamentSize)
	default:
		return fmt.Sprintf("toolbox.register('select', tools.%s)\n", gp.SelectionFunction)
	}
}

func (gp *GP) crossoverFunction() string {
	switch gp.CrossoverFunction {
	case "cxOnePoint":
		return "toolbox.register('mate', gp.cxOnePoint)\n"
	case "cxOnePointLeafBiased":
		return fmt.Sprintf("toolbox.register('mate', gp.cxOnePointLeafBiased, termpb=%v)\n", gp.TerminalProb)
	case "cxSemantic":
		return "toolbox.register('mate', gp.cxSemantic, gen_func=gp.genFull, pset=pset)\n"
	default:
		return "toolbox.register('mate', gp.cxOnePoint)\n"
	}
}

func (gp *GP) mutationFunction() string {
	switch gp.MutationFunction {
	case "mutUniform":
		return "toolbox.register('mutate', gp.mutUniform, expr=toolbox.expr_mut, pset=pset)\n"
	case "mutShrink":
		return "toolbox.register('mutate', gp.mutShrink)\n"
	case "mutNodeReplacement":
		return "toolbox.register('mutate', gp.mutNodeReplacement, pset=pset)\n"
	case "mutInsert":
		return "toolbox.register('mutate', gp.mutInsert, pset=pset)\n"
	case "mutEphemeral":
		return fmt.Sprintf("toolbox.register('mutate', gp.mutEphemeral, mode='%s')\n", gp.MutationMode)
	case "mutSemantic":
		return "toolbox.register('mutate', gp.mutSemantic, gen_func=gp.genFull, pset=pset)\n"
	default:
		return "toolbox.register('mutate', gp.mutUniform, expr=toolbox.expr_mut, pset=pset)\n"
	}
}

func (gp *GP) bloatControl() string {
	var code string
	code += fmt.Sprintf("toolbox.decorate('mate', gp.staticLimit(key=operator.attrgetter('height'), max_value=%d))\n", gp.MateHeight)
	code += fmt.Sprintf("toolbox.decorate('mutate', gp.staticLimit(key=operator.attrgetter('height'), max_value=%d))\n", gp.MutHeight)
	return code
}

func (gp *GP) setupStats() string {
	var code string
	code += "\tstats_fit = tools.Statistics(lambda ind: ind.fitness.values)\n"
	code += "\tstats_size = tools.Statistics(len)\n"
	code += "\tmstats = tools.MultiStatistics(fitness=stats_fit, size=stats_size)\n"
	code += "\tmstats.register('avg', numpy.mean)\n"
	code += "\tmstats.register('std', numpy.std)\n"
	code += "\tmstats.register('min', numpy.min)\n"
	code += "\tmstats.register('max', numpy.max)\n"
	return code

}

func (gp *GP) callAlgo() string {
	var code string
	switch gp.Algorithm {
	case "eaSimple":
		code += fmt.Sprintf("\tpop, logbook = algorithms.%s(pop, toolbox, cxpb=%v, mutpb=%v, ngen=%d, stats=mstats, halloffame=hof, verbose=True)\n", gp.Algorithm, gp.Cxpb, gp.Mutpb, gp.Generations)
	case "eaMuPlusLambda":
		code += fmt.Sprintf("\tpop, logbook = algorithms.%s(pop, toolbox, mu=%d, lambda_=%d, cxpb=%v, mutpb=%v, ngen=%d, stats=mstats, halloffame=hof, verbose=True)\n", gp.Algorithm, gp.Mu, gp.Lambda, gp.Cxpb, gp.Mutpb, gp.Generations)
	case "eaMuCommaLambda":
		code += fmt.Sprintf("\tpop, logbook = algorithms.%s(pop, toolbox, mu=%d, lambda_=%d, cxpb=%v, mutpb=%v, ngen=%d, stats=mstats, halloffame=hof, verbose=True)\n", gp.Algorithm, gp.Mu, gp.Lambda, gp.Cxpb, gp.Mutpb, gp.Generations)
	case "eaGenerateUpdate":
		code += "\tnumpy.random.seed(128)\n"
		code += fmt.Sprintf("\tstrategy = cma.Strategy(centroid=[5.0] * %d, sigma=5.0, lambda_=20 * %d)\n", gp.IndividualSize, gp.IndividualSize)
		code += "\ttoolbox.register('generate', strategy.generate, creator.Individual)\n"
		code += "\ttoolbox.register('update', strategy.update)\n"
		code += fmt.Sprintf("\tpop, logbook = algorithms.%s(toolbox, ngen=%d, stats=mstats, halloffame=hof, verbose=True)\n", gp.Algorithm, gp.Generations)
	default:
		code += fmt.Sprintf("\tpop, logbook = algorithms.%s(pop, toolbox, cxpb=%v, mutpb=%v, ngen=%d, stats=mstats, halloffame=hof, verbose=True)\n", gp.Algorithm, gp.Cxpb, gp.Mutpb, gp.Generations)
	}

	return code
}

func (gp *GP) setupLogs() string {
	var code string
	code += "\twith open(f\"{rootPath}/logbook.txt\", \"w\") as f:\n"
	code += "\t\tf.write(str(logbook))\n"
	code += "\n"

	// Write best individual to file.
	code += "\tout_file = open(f\"{rootPath}/best.txt\", \"w\")\n"
	code += "\tout_file.write(f\"Best individual fitness: {hof[0].fitness.values}\\n\")\n"
	code += "\tout_file.close()\n"
	return code
}

func (gp *GP) createPlots() string {
	var code string
	code += "\n\texpr = hof[0]\n"
	code += "\tnodes, edges, labels = gp.graph(expr)\n"
	code += "\tg = nx.Graph()\n"
	code += "\tg.add_nodes_from(nodes)\n"
	code += "\tg.add_edges_from(edges)\n"
	code += "\tpos = nx.nx_agraph.graphviz_layout(g, prog='dot')\n"
	code += "\n\tplt.figure(figsize=(7,7))\n"
	code += "\tnx.draw_networkx_nodes(g, pos, node_size=900, node_color='skyblue')\n"
	code += "\tnx.draw_networkx_edges(g, pos, edge_color='gray')\n"
	code += "\tnx.draw_networkx_labels(g, pos, labels, font_color='black')\n"
	code += "\tplt.axis('off')\n"
	code += "\tplt.savefig(f'{rootPath}/graph.png', dpi=300)\n"
	code += "\tplt.close()\n"
	return code
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
	code += "pset.addEphemeralConstant('rand101', partial(random.randint, -1, 1))\n"
	code += gp.renameArgs() + "\n"

	weights := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%f", gp.Weights), "[", "("), "]", ",)")
	code += fmt.Sprintf("creator.create('Fitness', base.Fitness, weights=%s)\n", weights)
	code += "creator.create('Individual', gp.PrimitiveTree, fitness=creator.Fitness)\n\n"

	code += fmt.Sprintf("toolbox.register('expr', gp.%s, pset=pset, min_=%d, max_=%d)\n", gp.Expr, gp.Min, gp.Max)
	code += fmt.Sprintf("toolbox.register('individual', tools.%s, creator.Individual, toolbox.expr)\n", gp.IndividualFunction)
	code += fmt.Sprintf("toolbox.register('population', tools.%s, list, toolbox.individual)\n", gp.PopulationFunction)
	code += "toolbox.register('compile', gp.compile, pset=pset)\n\n"

	code += fmt.Sprintf("toolbox.register('evaluate', evalSymbReg, points=[x / 10.0 for x in range(-10, 10)], realFunction=\"%v\")\n\n", gp.RealFunction)

	code += gp.selectionFunction() + "\n"
	code += gp.crossoverFunction() + "\n"
	code += fmt.Sprintf("toolbox.register('expr_mut', gp.%s, min_=%d, max_=%d)\n", gp.ExprMut, gp.ExprMutMin, gp.ExprMutMax)
	code += gp.mutationFunction() + "\n"
	code += gp.bloatControl() + "\n"
	code += "toolbox.register('map', futures.map)\n"

	code += "def main():\n"
	code += "\trootPath = os.path.dirname(os.path.abspath(__file__))\n"
	code += "\trandom.seed(318)\n"
	code += fmt.Sprintf("\tpop = toolbox.population(n=%d)\n", gp.PopulationSize)
	code += fmt.Sprintf("\thof = tools.HallOfFame(%d)\n", gp.HofSize)
	code += gp.setupStats() + "\n"
	code += "\tN = " + fmt.Sprintf("%d", gp.IndividualSize) + "\n"
	code += gp.callAlgo() + "\n"
	code += gp.setupLogs() + "\n"
	code += gp.createPlots() + "\n"

	code += "\n\n"
	code += "if __name__ == '__main__':\n"
	code += "\tmain()\n"

	return code, nil
}
