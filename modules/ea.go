package modules

import "evolve/util"

type EA struct {
	Algorithm          string    `json:"algorithm"`
	Individual         string    `json:"individual"`
	PopulationFunction string    `json:"populationFunction"`
	EvaluationFunction string    `json:"evaluationFunction"`
	PopulationSize     int       `json:"populationSize"`
	Generations        int       `json:"generations"`
	Cxpb               float64   `json:"cxpb"`
	Mutpb              float64   `json:"mutpb"`
	Weights            []float64 `json:"weights"`
	IndividualSize     int       `json:"individualSize"`
	Indpb              float64   `json:"indpb"`
	RandomRange        []float64 `json:"randomRange"`
	CrossoverFunction  string    `json:"crossoverFunction"`
	MutationFunction   string    `json:"mutationFunction"`
	SelectionFunction  string    `json:"selectionFunction"`
	TournamentSize     *int      `json:"tournamentSize,omitempty"`
	Mu                 *int      `json:"mu,omitempty"`
	Lambda_            *int      `json:"lambda,omitempty"`
	HofSize            *int      `json:"hofSize,omitempty"`
}

func (ea *EA) Validate() error {
	if err := util.ValidateAlgorithmName(ea.Algorithm); err != nil {
		return err
	}
	// TODO: Validate remaining fields.
	return nil
}
