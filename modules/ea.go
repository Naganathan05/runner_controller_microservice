package modules

import (
	"encoding/json"
	"evolve/util"
	"fmt"
	"slices"
	"strings"
)

type EA struct {
	Algorithm          string    `json:"algorithm"`
	Individual         string    `json:"individual"`
	PopulationFunction string    `json:"populationFunction"`
	CustomPop          string    `json:"customPop,omitempty"` // NEW
	EvaluationFunction string    `json:"evaluationFunction"`
	CustomEval         string    `json:"customEval,omitempty"` // NEW
	PopulationSize     int       `json:"populationSize"`
	Generations        int       `json:"generations"`
	Cxpb               float64   `json:"cxpb"`
	Mutpb              float64   `json:"mutpb"`
	Weights            []float64 `json:"weights"`
	IndividualSize     int       `json:"individualSize"` // Number of Dimensions.
	Indpb              float64   `json:"indpb"`
	RandomRange        []float64 `json:"randomRange"`
	CrossoverFunction  string    `json:"crossoverFunction"`
	MutationFunction   string    `json:"mutationFunction"`
	CustomMutation     string    `json:"customMutation,omitempty"` // NEW
	SelectionFunction  string    `json:"selectionFunction"`
	CustomSelection    string    `json:"customSelection,omitempty"` // NEW
	TournamentSize     int       `json:"tournamentSize,omitempty"`
	Mu                 int       `json:"mu,omitempty"`
	Lambda             int       `json:"lambda_,omitempty"`
	HofSize            int       `json:"hofSize,omitempty"`

	// Differential Evolution Params.
	CrossOverRate float64 `json:"crossOverRate,omitempty"`
	ScalingFactor float64 `json:"scalingFactor,omitempty"`
}

func EAFromJSON(jsonData map[string]any) (*EA, error) {
	ea := &EA{}
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonDataBytes, ea); err != nil {
		return nil, err
	}
	return ea, nil
}

func (ea *EA) validate() error {
	if err := util.ValidateAlgorithmName(ea.Algorithm); err != nil {
		return err
	}
	// TODO: Validate remaining fields.
	return nil
}

func (ea *EA) imports() string {
	return strings.Join([]string{
		"import random",
		"from deap import base, creator, tools, algorithms",
		"import numpy",
		"import matplotlib.pyplot as plt",
		"from functools import reduce",
		"from scoop import futures",
		"from deap import benchmarks",
	}, "\n")
}

// If the function is a built-in function, return the corresponding Python code.
// Otherwise, return the function string as is.
func (ea *EA) evalFunction() string {
	if slices.Contains([]string{"rand", "plane", "sphere", "cigar", "rosenbrock", "h1", "ackley", "bohachevsky", "griewank", "rastrigin", "rastrigin_scaled", "rastrigin_skew", "schaffer", "schwefel", "himmelblau"}, ea.EvaluationFunction) {
		ea.EvaluationFunction = "benchmarks." + ea.EvaluationFunction
		return ""
	}

	switch ea.EvaluationFunction {
	case "evalOneMax":
		return "def evalOneMax(individual):\n    return sum(individual),"
	case "evalProduct":
		return "def evalProduct(individual):\n    return reduce(lambda x, y: x*y, individual),"
	case "evalDifference":
		return "def evalDifference(individual):\n    return reduce(lambda x, y: x-y, individual),"
	default:
		return ea.CustomEval
	}
}

func (ea *EA) registerIndividual() string {
	// TODO: Add support for string individual types with initial seed.
	switch strings.ToLower(ea.Individual) {
	case "binarystring":
		return "toolbox.register(\"attr\", random.randint, 0, 1)\n"
	case "floatingpoint":
		return fmt.Sprintf("toolbox.register(\"attr\", random.uniform, %f, %f)\n", ea.RandomRange[0], ea.RandomRange[1])
	case "integer":
		return fmt.Sprintf("toolbox.register(\"attr\", random.randint, %f, %f)\n", ea.RandomRange[0], ea.RandomRange[1])
	default:
		return ""
	}
}

func (ea *EA) initialGenerator() string {
	// TODO: Add support for other generator functions.
	switch ea.PopulationFunction {
	case "initRepeat":
		return fmt.Sprintf("toolbox.register(\"individual\", tools.%s, creator.Individual, toolbox.attr, %d)\n", ea.PopulationFunction, ea.IndividualSize) + fmt.Sprintf("toolbox.register(\"population\", tools.%s, list, toolbox.individual)\n", ea.PopulationFunction)
	default:
		return ea.PopulationFunction
	}
}

func (ea *EA) mutationFunction() string {
	// TODO: Add support for more mutation functions.
	switch ea.MutationFunction {
	case "mutFlipBit":
		return fmt.Sprintf("toolbox.register(\"mutate\", tools.%s, indpb=%f)\n", ea.MutationFunction, ea.Indpb)
	case "mutShuffleIndexes":
		return fmt.Sprintf("toolbox.register(\"mutate\", tools.%s, indpb=%f)\n", ea.MutationFunction, ea.Indpb)
	default:
		return ea.MutationFunction
	}
}

func (ea *EA) selectionFunction() string {
	// TODO: Add support for more selection functions.
	switch ea.SelectionFunction {
	case "selTournament":
		return fmt.Sprintf("toolbox.register(\"select\", tools.%s, tournsize=%d)\n", ea.SelectionFunction, ea.TournamentSize)
	default:
		return fmt.Sprintf("toolbox.register(\"select\", tools.%s)\n", ea.SelectionFunction)
	}
}

func (ea *EA) callAlgo() string {
	switch strings.ToLower(ea.Algorithm) {
	case "easimple":
		return "\tpop, logbook = algorithms.eaSimple(pop, toolbox, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"
	case "eamupluslambda":
		return fmt.Sprintf("\tmu = %d\n", ea.Mu) + fmt.Sprintf("\tlambda_ = %d\n", ea.Lambda) + "\tpop, logbook = algorithms.eaMuPlusLambda(pop, toolbox, mu=mu, lambda_=lambda_, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"
	case "eamucommalambda":
		return fmt.Sprintf("\tmu = %d\n", ea.Mu) + fmt.Sprintf("\tlambda_ = %d\n", ea.Lambda) + "\tpop, logbook = algorithms.eaMuCommaLambda(pop, toolbox, mu=mu, lambda_=lambda_, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"
	case "eagenerateupdate":
		return fmt.Sprintf("\tstrategy = cma.Strategy(centroid=[5.0]*%d, sigma=5.0, lambda_=20*{N})\n", ea.IndividualSize) + "\ttoolbox.register(\"generate\", strategy.generate, creator.Individual)\n" + "\ttoolbox.register(\"update\", strategy.update)\n" + "\tpop, logbook = algorithms.eaGenerateUpdate(toolbox, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"
	default:
		return "\tpop, logbook = algorithms.eaSimple(pop, toolbox, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"
	}
}

func (ea *EA) plots() string {
	var plots string

	// Fitness Plot.
	plots += "\n\n"
	plots += "\tgen = logbook.select(\"gen\")\n"
	plots += "\tavg = logbook.select(\"avg\")\n"
	plots += "\tmin_ = logbook.select(\"min\")\n"
	plots += "\tmax_ = logbook.select(\"max\")\n\n"
	plots += "\tplt.plot(gen, avg, label=\"average\")\n"
	plots += "\tplt.plot(gen, min_, label=\"minimum\")\n"
	plots += "\tplt.plot(gen, max_, label=\"maximum\")\n"
	plots += "\tplt.xlabel(\"Generation\")\n"
	plots += "\tplt.ylabel(\"Fitness\")\n"
	plots += "\tplt.legend(loc=\"lower right\")\n"
	plots += "\tplt.savefig(f\"fitness_plot.png\", dpi=300)\n"
	plots += "\tplt.close()\n"
	plots += "\n\n"

	// Mutation and Crossover Effect.
	plots += "\tavg_fitness = logbook.select(\"avg\")\n"
	plots += "\tfitness_diff = [avg_fitness[i] - avg_fitness[i-1] for i in range(1, len(avg_fitness))]\n"
	plots += "\tplt.plot(gen[1:], fitness_diff, label=\"Fitness Change\", color=\"purple\")\n"
	plots += "\tplt.xlabel(\"Generation\")\n"
	plots += "\tplt.ylabel(\"Fitness Change\")\n"
	plots += "\tplt.title(\"Effect of Mutation and Crossover on Fitness\")\n"
	plots += "\tplt.legend()\n"
	plots += "\tplt.savefig(f\"mutation_crossover_effect.png\", dpi=300)\n"
	plots += "\tplt.close()\n"
	return plots
}

func (ea *EA) differentialEvolution() string {
	return strings.Join([]string{
		fmt.Sprintf("\tCR = %f", ea.CrossOverRate),
		fmt.Sprintf("F = %f", ea.ScalingFactor),
		"\n",
		"logbook = tools.Logbook()",
		"logbook.header = 'gen', 'evals', 'std', 'min', 'avg', 'max'",
		"fitnesses = toolbox.map(toolbox.evaluate, pop)",
		"for ind, fit in zip(pop, fitnesses):",
		"\tind.fitness.values = fit",
		"record = stats.compile(pop)",
		"logbook.record(gen=0, evals=len(pop), **record)",
		"print(logbook.stream)",
		"for g in range(1, generations):",
		"\tfor k, agent in enumerate(pop):",
		"\t\ta,b,c = toolbox.select(pop, 3)",
		"\t\ty = toolbox.clone(agent)",
		"\t\tindex = random.randrange(N)",
		"\t\tfor i, value in enumerate(agent):",
		"\t\t\tif i == index or random.random() < CR:",
		"\t\t\t\ty[i] = a[i] + F*(b[i]-c[i])",
		"\t\ty.fitness.values = toolbox.evaluate(y)",
		"\t\tif y.fitness > agent.fitness:",
		"\t\t\tpop[k] = y",
		"\thof.update(pop)",
		"\trecord = stats.compile(pop)",
		"\tlogbook.record(gen=g, evals=len(pop), **record)",
		"\tprint(logbook.stream)",
		"\n",
	}, "\n\t")
}

func (ea *EA) Code() (string, error) {
	if err := ea.validate(); err != nil {
		return "", err
	}

	var code string
	code += ea.imports() + "\n\n"
	code += ea.evalFunction() + "\n\n"
	code += ea.CustomPop + "\n"
	code += ea.CustomMutation + "\n"
	code += ea.CustomSelection + "\n\n"

	code += "toolbox = base.Toolbox()\n\n"
	weights := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%f", ea.Weights), "[", "("), "]", ",)")
	code += fmt.Sprintf("creator.create('FitnessMax', base.Fitness, weights=%s)\n", weights)
	code += "creator.create(\"Individual\", list, fitness=creator.FitnessMax)\n\n"

	code += ea.registerIndividual() + "\n"
	code += ea.initialGenerator() + "\n"
	code += fmt.Sprintf("toolbox.register(\"evaluate\", %s)\n", ea.EvaluationFunction)
	code += ea.mutationFunction() + "\n"
	code += fmt.Sprintf("toolbox.register(\"mate\", tools.%s)\n", ea.CrossoverFunction)
	code += ea.selectionFunction() + "\n"
	code += "\ntoolbox.register(\"map\", futures.map)\n\n"

	code += "def main():\n"
	code += fmt.Sprintf("\tpopulationSize = %d\n", ea.PopulationSize)
	code += fmt.Sprintf("\tgenerations = %d\n", ea.Generations)
	code += fmt.Sprintf("\tcxpb = %f\n", ea.Cxpb)
	code += fmt.Sprintf("\tmutpb = %f\n", ea.Mutpb)
	code += fmt.Sprintf("\tN = %d\n", ea.IndividualSize)
	code += "\n\tpop = toolbox.population(n=populationSize)\n"
	code += fmt.Sprintf("\thof = tools.HallOfFame(%d)\n", ea.HofSize)
	code += "\n\tstats = tools.Statistics(lambda ind: ind.fitness.values)\n"
	code += "\tstats.register(\"avg\", numpy.mean)\n"
	code += "\tstats.register(\"min\", numpy.min)\n"
	code += "\tstats.register(\"max\", numpy.max)\n"
	code += "\n"

	if ea.Algorithm == "de" {
		code += ea.differentialEvolution()
	} else {
		code += ea.callAlgo() + "\n"
	}

	code += "\tprint(f'Best individual is: {hof[0]}\\nwith fitness: {hof[0].fitness}')"
	code += "\n\n"
	code += ea.plots()
	code += "\n\n"
	code += "if __name__ == '__main__':\n"
	code += "\tmain()"

	return code, nil
}
