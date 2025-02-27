package modules

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
