package modules

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
	TournamentSize           int      `json:"tournamentSize,omitempty"`
	Mu                       int      `json:"mu,omitempty"`
	Lambda_                  int      `json:"lambda,omitempty"`
	HofSize                  int      `json:"hofSize,omitempty"`
}
