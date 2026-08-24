package sample

// State is the console-facing sampler snapshot.
type State struct {
	Rate     int `json:"rate"`
	Capacity int `json:"capacity"`
	Token    int `json:"token"`
	Sampled  int `json:"sampled"`
	Rejected int `json:"rejected"`
	Budget   int `json:"budget"`
}
