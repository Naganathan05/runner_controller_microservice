package modules

import (
	"encoding/json"
	"evolve/util"
	"fmt"
	"strings"
)

type EAML struct {
	Algorithm                string    `json:"algorithm"`
	MlEvalFunctionCodeString string    `json:"mlEvalFunctionCodeString"`
	PopulationSize           int       `json:"populationSize"`
	Generations              int       `json:"generations"`
	Cxpb                     float64   `json:"cxpb"`
	Mutpb                    float64   `json:"mutpb"`
	Weights                  []float64 `json:"weights"`
	GoogleDriveUrl           string    `json:"googleDriveUrl"`
	Sep                      string    `json:"sep"`
	MlImportCodeString       string    `json:"mlImportCodeString"`
	TargetColumnName         string    `json:"targetColumnName"`
	Indpb                    float64   `json:"indpb"`
	CrossoverFunction        string    `json:"crossoverFunction"`
	MutationFunction         string    `json:"mutationFunction"`
	SelectionFunction        string    `json:"selectionFunction"`
	TournamentSize           int       `json:"tournamentSize,omitempty"`
	Mu                       int       `json:"mu,omitempty"`
	Lambda                   int       `json:"lambda_,omitempty"`
	HofSize                  int       `json:"hofSize,omitempty"`
}

func MLFromJSON(jsonData map[string]any) (*EAML, error) {
	ml := &EAML{}
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonDataBytes, ml); err != nil {
		return nil, err
	}
	return ml, nil
}

func (ml *EAML) validate() error {
	if err := util.ValidateAlgorithmName(ml.Algorithm); err != nil {
		return err
	}
	// TODO: Validate remaining fields.
	return nil
}

func (ml *EAML) imports() string {
	return strings.Join([]string{
		"# DEAP imports",
		"import random, os",
		"from deap import base, creator, tools, algorithms",
		"import numpy",
		"import matplotlib.pyplot as plt",
		"from functools import reduce",
		"from scoop import futures",
		"import pandas as pd",
		"import warnings",
		"warnings.filterwarnings(\"ignore\")",
		"# ML imports",
		ml.MlImportCodeString,
	}, "\n")
}

func (ml *EAML) googleDriveDownloadFunc() string {
	return strings.Join([]string{
		"def download_csv_from_google_drive_share_link(url):",
		"\tfile_id = url.split(\"/\")[-2]",
		"\tdwn_url = \"https://drive.google.com/uc?export=download&id=\" + file_id",
		"\treturn pd.read_csv(dwn_url, sep=\"" + ml.Sep + "\")\n",
	}, "\n")
}

func (ml *EAML) selectionFunction() string {
	// TODO: Add support for other selection functions.
	switch ml.SelectionFunction {
	case "selTournament":
		return fmt.Sprintf("toolbox.register('select', tools.%s, tournsize=%d)\n", ml.SelectionFunction, ml.TournamentSize)
	default:
		return fmt.Sprintf("toolbox.register('select', tools.%s)\n", ml.SelectionFunction)
	}
}

func (ml *EAML) callAlgo() string {
	switch ml.Algorithm {
	case "eaSimple":
		return "\tpop, logbook = algorithms.eaSimple(pop, toolbox, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"

	case "eaMuPlusLambda":
		return fmt.Sprintf("\tmu = %d\n", ml.Mu) + fmt.Sprintf("\tlambda_ = %d\n", ml.Lambda) + "\tpop, logbook = algorithms.eaMuPlusLambda(pop, toolbox, mu=mu, lambda_=lambda_, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"

	case "eaMuCommaLambda":
		return fmt.Sprintf("\tmu = %d\n", ml.Mu) + fmt.Sprintf("\tlambda_ = %d\n", ml.Lambda) + "\tpop, logbook = algorithms.eaMuCommaLambda(pop, toolbox, mu=mu, lambda_=lambda_, cxpb=cxpb, mutpb=mutpb, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"

	case "eaGenerateUpdate":
		return "\tnumpy.random.seed(128)\n" + fmt.Sprintf("\tstrategy = cma.Strategy(centroid=[5.0]*len(X.columns), sigma=5.0, lambda_=%d*len(X.columns))\n", ml.Lambda) + "\ttoolbox.register(\"generate\", strategy.generate, creator.Individual)\n" + "\ttoolbox.register(\"update\", strategy.update)\n" + "\tpop, logbook = algorithms.eaGenerateUpdate(toolbox, ngen=generations, stats=stats, halloffame=hof, verbose=True)\n"

	default:
		return ""
	}
}

func (ml *EAML) createPlots() string {
	return strings.Join([]string{
		"\n\n",
		"\tgen = logbook.select(\"gen\")",
		"\tavg = logbook.select(\"avg\")",
		"\tmin_ = logbook.select(\"min\")",
		"\tmax_ = logbook.select(\"max\")\n",

		// Save LogBook as .log.
		"\twith open(f\"{rootPath}/logbook.txt\", \"w\") as f:",
		"\t\tf.write(str(logbook))",
		"\n",

		// Save fitness plot.
		"\tplt.plot(gen, avg, label=\"average\")",
		"\tplt.plot(gen, min_, label=\"minimum\")",
		"\tplt.plot(gen, max_, label=\"maximum\")",
		"\tplt.xlabel(\"Generation\")",
		"\tplt.ylabel(\"Fitness\")",
		"\tplt.legend(loc=\"lower right\")",
		"\tplt.savefig(f\"{rootPath}/fitness_plot.png\", dpi=300)",
		"\tplt.close()",
	}, "\n")
}

func (ml *EAML) Code() (string, error) {
	if err := ml.validate(); err != nil {
		return "", err
	}

	var code string
	code += ml.imports() + "\n"
	code += ml.googleDriveDownloadFunc() + "\n"
	code += ml.MlEvalFunctionCodeString + "\n"

	code += "toolbox = base.Toolbox()\n\n"
	code += "toolbox.register(\"mate\", tools." + ml.CrossoverFunction + ")\n"
	code += fmt.Sprintf("toolbox.register(\"mutate\", tools.%v, indpb=%v)\n", ml.MutationFunction, ml.Indpb)
	code += ml.selectionFunction() + "\n"
	code += "\ntoolbox.register(\"map\", futures.map)\n\n"

	weights := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%f", ml.Weights), "[", "("), "]", ",)")
	code += strings.Join([]string{
		"def main():",
		"\trootPath = os.path.dirname(os.path.abspath(__file__))",
		fmt.Sprintf("\turl = \"%s\"", ml.GoogleDriveUrl),
		"\tdf = download_csv_from_google_drive_share_link(url)",
		fmt.Sprintf("\ttarget = \"%s\"", ml.TargetColumnName),
		"\tX = df.drop(target, axis=1)",
		"\ty = df[target]",
		"\taccuracy = mlEvalFunction([1 for _ in range(len(X.columns))], X, y)",
		fmt.Sprintf("\tcreator.create(\"FitnessMax\", base.Fitness, weights=%v)", weights),
		"\tcreator.create(\"Individual\", list, fitness=creator.FitnessMax)",
		"\ttoolbox.register(\"attr\", random.randint, 0, 1)",
	}, "\n") + "\n"

	code += "\ttoolbox.register(\"individual\", tools.initRepeat, creator.Individual, toolbox.attr, n=len(X.columns))\n"
	code += "\ttoolbox.register(\"population\", tools.initRepeat, list, toolbox.individual)\n"

	code += "\ttoolbox.register(\"evaluate\", mlEvalFunction, X=X, y=y)\n"

	code += fmt.Sprintf("\tpopulationSize = %d\n", ml.PopulationSize)
	code += fmt.Sprintf("\tgenerations = %d\n", ml.Generations)
	code += fmt.Sprintf("\tcxpb = %v\n", ml.Cxpb)
	code += fmt.Sprintf("\tmutpb = %v\n", ml.Mutpb)
	code += "\tN = len(X.columns)\n"
	code += fmt.Sprintf("\thofSize = %d\n", ml.HofSize)
	code += "\n\tpop = toolbox.population(n=populationSize)\n"
	code += "\thof = tools.HallOfFame(hofSize)\n"
	code += "\n\tstats = tools.Statistics(lambda ind: ind.fitness.values)\n"
	code += "\tstats.register(\"avg\", numpy.mean)\n"
	code += "\tstats.register(\"min\", numpy.min)\n"
	code += "\tstats.register(\"max\", numpy.max)\n"

	code += ml.callAlgo()
	code += "\tout_file = open(f\"{rootPath}/best.txt\", \"w\")\n"
	code += "\tout_file.write(f\"Before applying EA: {accuracy}\\n\")\n"
	code += "\tout_file.write(f\"Best individual is:\\n{hof[0]}\\nwith fitness: {hof[0].fitness}\\n\")\n"
	code += "\tbest_columns = [i for i in range(len(hof[0])) if hof[0][i] == 1]\n"
	code += "\tbest_column_names = X.columns[best_columns]\n"
	code += "\tout_file.write(f\"\\nBest individual columns:\\n{best_column_names.values}\")\n"
	code += "\tout_file.close()\n"

	code += ml.createPlots()
	code += "\n\n"
	code += "if __name__ == '__main__':\n"
	code += "\tmain()\n"

	return code, nil
}
