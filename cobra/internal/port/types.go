package port

type PortResult struct {
	Host  string
	Port  int
	Open  bool
	Error error
}
