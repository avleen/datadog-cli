package modules

type Module interface {
	Name() string
	ParseFlags(args []string) error
}
