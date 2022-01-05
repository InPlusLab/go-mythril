package z3

// #include <stdlib.h>
// #cgo CFLAGS: -IC:/Z3/src/api
// #cgo LDFLAGS: -LC:/Z3/build -llibz3
// #include "z3.h"
import "C"

// LBool is the lifted boolean type representing false, true, and undefined.
type LBool int8

const (
	False LBool = C.Z3_L_FALSE
	Undef       = C.Z3_L_UNDEF
	True        = C.Z3_L_TRUE
)
