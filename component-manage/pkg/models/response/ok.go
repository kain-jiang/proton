package response

type Status struct {
	Status string `json:"status"`
}

var OK = Status{Status: "ok"}
