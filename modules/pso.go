package modules

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type PSO struct {
	Algorithm      string    `json:"algorithm"` // Original PSO, MultiSwarm, Speciation.
	Weights        []float64 `json:"weights"`
	Dimensions     int       `json:"dimensions"`
	MinPosition    float64   `json:"minPosition"`
	MaxPosition    float64   `json:"maxPosition"`
	MinSpeed       float64   `json:"minSpeed"`
	MaxSpeed       float64   `json:"maxSpeed"`
	Phi1           float64   `json:"phi1"`      // cognitive component
	Phi2           float64   `json:"phi2"`      // social component
	Benchmark      string    `json:"benchmark"` // Evaluation function.
	PopulationSize int       `json:"populationSize"`
	Generations    int       `json:"generations"`
}

func PSOFromJSON(jsonData map[string]any) (*PSO, error) {
	pso := &PSO{}
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonDataBytes, pso); err != nil {
		return nil, err
	}
	return pso, nil
}

func (pso *PSO) validate() error {
	if !slices.Contains([]string{"original", "multiswarm", "speciation"}, pso.Algorithm) {
		return fmt.Errorf("invalid PSO algorithm name: %s", pso.Algorithm)
	}

	if pso.Dimensions <= 0 {
		return fmt.Errorf("invalid number of dimensions: %d", pso.Dimensions)
	}

	if pso.MinPosition >= pso.MaxPosition {
		return fmt.Errorf("invalid min/max position: %f/%f", pso.MinPosition, pso.MaxPosition)
	}

	if pso.MinSpeed >= pso.MaxSpeed {
		return fmt.Errorf("invalid min/max speed: %f/%f", pso.MinSpeed, pso.MaxSpeed)
	}

	if !slices.Contains([]string{"rand", "plane", "sphere", "cigar", "rosenbrock", "h1", "ackley", "bohachevsky", "griewank", "rastrigin", "rastrigin_scaled", "rastrigin_skew", "schaffer", "schwefel", "himmelblau"}, pso.Benchmark) {
		return fmt.Errorf("invalid benchmark function: %s", pso.Benchmark)
	}

	if pso.PopulationSize <= 0 {
		return fmt.Errorf("invalid population size: %d", pso.PopulationSize)
	}

	if pso.Generations <= 0 {
		return fmt.Errorf("invalid number of generations: %d", pso.Generations)
	}

	return nil
}

func (pso *PSO) imports() string {
	return strings.Join([]string{
		"import math, os",
		"import numpy",
		"from deap import base, benchmarks, creator, tools",
		"import matplotlib.pyplot as plt",
		"import matplotlib.animation as animation",
	}, "\n")
}

func (pso *PSO) generateAndUpdateParticle() string {
	return strings.Join([]string{
		"def generate(size, pmin, pmax, smin, smax):",
		"\tpart = creator.Particle(numpy.random.uniform(pmin, pmax, size))",
		"\tpart.speed = numpy.random.uniform(smin, smax, size)",
		"\tpart.smin = smin",
		"\tpart.smax = smax",
		"\treturn part\n",
		"def updateParticle(part, best, phi1, phi2):",
		"\tu1 = numpy.random.uniform(0, phi1, len(part))",
		"\tu2 = numpy.random.uniform(0, phi2, len(part))",
		"\tv_u1 = u1 * (part.best - part)",
		"\tv_u2 = u2 * (best - part)",
		"\tpart.speed += v_u1 + v_u2",
		"\tfor i, speed in enumerate(part.speed):",
		"\t\tif abs(speed) < part.smin:",
		"\t\t\tpart.speed[i] = math.copysign(part.smin, speed)",
		"\t\telif abs(speed) > part.smax:",
		"\t\t\tpart.speed[i] = math.copysign(part.smax, speed)",
		"\tpart += part.speed",
	}, "\n")
}

func (pso *PSO) toolbox() string {
	return strings.Join([]string{
		"toolbox = base.Toolbox()",
		fmt.Sprintf("toolbox.register('particle', generate, size=%d, pmin=%f, pmax=%f, smin=%f, smax=%f)", pso.Dimensions, pso.MinPosition, pso.MaxPosition, pso.MinSpeed, pso.MaxSpeed),
		"toolbox.register('population', tools.initRepeat, list, toolbox.particle)",
		fmt.Sprintf("toolbox.register('update', updateParticle, phi1=%f, phi2=%f)", pso.Phi1, pso.Phi2),
		fmt.Sprintf("toolbox.register('evaluate', benchmarks.%s)", pso.Benchmark),
	}, "\n")
}

func (pso *PSO) setupPlot() string {
	return strings.Join([]string{
		"\n\tfig, ax = plt.subplots()",
		"\t# Dynamically determine plot space based on initial particle positions.",
		"\tall_x = [p[0] for p in pop]",
		"\tall_y = [p[1] for p in pop]",
		"\tx_min = min(all_x)",
		"\tx_max = max(all_x)",
		"\ty_min = min(all_y)",
		"\ty_max = max(all_y)",
		"\t# Add a buffer to the plot limits to ensure particles don't get cut off",
		"\tx_buffer = (x_max - x_min) * 0.5",
		"\ty_buffer = (y_max - y_min) * 0.5",
		"\tx_min -= x_buffer",
		"\tx_max += x_buffer",
		"\ty_min -= y_buffer",
		"\ty_max += y_buffer",
		"\tax.set_xlim(x_min, x_max)",
		"\tax.set_ylim(y_min, y_max)",
		"\tscat = ax.scatter([p[0] for p in pop], [p[1] for p in pop])  # Initial scatter plot",
		"\tbest_scat = ax.scatter([], [], color='red', marker='*', s=100) # Scatter plot for the best particle",
		"\tplt.xlabel('x')",
		"\tplt.ylabel('y')",
		"\tplt.title('Particle Swarm Optimization')",
		"\tgeneration_text = ax.text(0.02, 0.95, '', transform=ax.transAxes)  # Text to display generation",
	}, "\n")
}

func (pso *PSO) thePSOAlgo() string {
	return strings.Join([]string{
		"\tdef update(frame):",
		"\tnonlocal best, pop, x_min, x_max, y_min, y_max  # Access the pop variable and plot limits",
		"\tfor part in pop:",
		"\t\tpart.fitness.values = toolbox.evaluate(part)",
		"\t\tif part.best is None or part.best.fitness < part.fitness:",
		"\t\t\tpart.best = creator.Particle(part)",
		"\t\t\tpart.best.fitness.values = part.fitness.values",
		"\t\tif best is None or best.fitness < part.fitness:",
		"\t\t\tbest = creator.Particle(part)",
		"\t\t\tbest.fitness.values = part.fitness.values",
		"\tfor part in pop:",
		"\t\ttoolbox.update(part, best)",
		"\t# Update scatter plot positions",
		"\tscat.set_offsets(numpy.array([[p[0], p[1]] for p in pop]))",
		"\t# Update best particle position",
		"\tbest_scat.set_offsets(numpy.array([[best[0], best[1]]]))",
		"\t# Update generation text",
		"\tgeneration_text.set_text(f'Generation: {frame}')",
		"\t# Dynamically adjust the plot limits based on particle positions",
		"\tall_x = [p[0] for p in pop]",
		"\tall_y = [p[1] for p in pop]",
		"\tcurr_x_min = min(all_x)",
		"\tcurr_x_max = max(all_x)",
		"\tcurr_y_min = min(all_y)",
		"\tcurr_y_max = max(all_y)",
		"\tx_buffer = (curr_x_max - curr_x_min) * 0.5",
		"\ty_buffer = (curr_y_max - curr_y_min) * 0.5",
		"\tcurr_x_min -= x_buffer",
		"\tcurr_x_max += x_buffer",
		"\tcurr_y_min -= y_buffer",
		"\tcurr_y_max += y_buffer",
		"\t# Expand the plot space if particles are going out of range, but never shrink it.",
		"\tif curr_x_min < x_min:",
		"\t\tx_min = curr_x_min",
		"\tif curr_x_max > x_max:",
		"\t\tx_max = curr_x_max",
		"\tif curr_y_min < y_min:",
		"\t\ty_min = curr_y_min",
		"\tif curr_y_max > y_max:",
		"\t\ty_max = curr_y_max",
		"\tax.set_xlim(x_min, x_max)",
		"\tax.set_ylim(y_min, y_max)",
		"\t# Gather all the fitnesses in one list and print the stats",
		"\tlogbook.record(gen=frame, evals=len(pop), **stats.compile(pop))",
		"\tprint(logbook.stream)",
		"\treturn scat, best_scat, generation_text",
	}, "\n\t")
}

func (pso *PSO) Code() (string, error) {
	if err := pso.validate(); err != nil {
		return "", err
	}

	var code string
	code += pso.imports() + "\n\n"
	weights := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%f", pso.Weights), "[", "("), "]", ",)")
	code += fmt.Sprintf("creator.create('FitnessMax', base.Fitness, weights=%s)\n", weights)
	code += "creator.create('Particle', numpy.ndarray, fitness=creator.FitnessMax, speed=list, smin=None, smax=None, best=None)\n\n"
	code += pso.generateAndUpdateParticle() + "\n"
	code += pso.toolbox() + "\n"

	// main
	code += strings.Join([]string{
		"def main():",
		"\trootPath = os.path.dirname(os.path.abspath(__file__))",
		fmt.Sprintf("\tpop = toolbox.population(n=%d)", pso.PopulationSize),
		"\tstats = tools.Statistics(lambda ind: ind.fitness.values)",
		"\tstats.register('avg', numpy.mean)",
		"\tstats.register('std', numpy.std)",
		"\tstats.register('min', numpy.min)",
		"\tstats.register('max', numpy.max)",
		"\n\tlogbook = tools.Logbook()",
		"\tlogbook.header = ['gen', 'evals'] + stats.fields",
		"\n\tbest = None",
		fmt.Sprintf("\tGEN = %d", pso.Generations),
	}, "\n")

	code += pso.setupPlot() + "\n"
	code += pso.thePSOAlgo() + "\n"

	code += strings.Join([]string{
		"\tani = animation.FuncAnimation(fig, update, frames=GEN, blit=True, repeat=False)",
		"\tani.save(f'{rootPath}/pso_animation.gif', writer='pillow', fps=10)",

		"\t# Save the position of the best particle",
		"\tout_file = open(f'{rootPath}/best.txt', 'w')",
		"\tout_file.write(f'Best individual fitness: {best.fitness.values}\\n')",
		"\tout_file.write(f'Best individual position: {best}\\n')",
		"\tout_file.close()",

		"\twith open(f'{rootPath}/logbook.txt', 'w') as f:",
		"\t\tf.write(str(logbook))",

		"if __name__ == '__main__':",
		"\tmain()",
	}, "\n")

	return code, nil
}
