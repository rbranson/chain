package iface

type Unwrap interface {
	Unwrap() (interface{}, bool)
}

type Is interface {
	Is(other interface{}) bool
}

type As interface {
	As(other interface{}) bool
}

type Wrap interface {
	Wrap(v interface{}) bool
}
