package ext

type Inheritable interface {
	Inherit() Inheritable
}
