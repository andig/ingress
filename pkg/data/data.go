package data

type Data struct {
	ID    string
	Name  string
	Value float64
}

type InputData struct {
	*Data
	Source string
}
