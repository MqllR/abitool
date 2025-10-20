package contract

type Printer interface {
	Print() (string, error)
}

func NoopPrinter(_ string) (string, error) {
	return "", nil
}
