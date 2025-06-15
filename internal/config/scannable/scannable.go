package scannable

type Scannable struct {
	Sender  string
	Headers map[string]string
	Body    []byte
}
