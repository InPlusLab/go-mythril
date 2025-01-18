package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

type Statistics struct {
	RawCtx C.Z3_context
	Stats  C.Z3_stats
}

func (s *Statistics) GetKeyValue(key string) int {
	for idx := 0; idx < s.Length(); idx++ {
		//fmt.Println(idx, C.Z3_stats_get_key(s.RawCtx, s.Stats, C.uint(idx)), C.GoString(C.Z3_stats_get_key(s.RawCtx, s.Stats, C.uint(idx))))
		if key == C.GoString(C.Z3_stats_get_key(s.RawCtx, s.Stats, C.uint(idx))) {
			if bool(C.Z3_stats_is_uint(s.RawCtx, s.Stats, C.uint(idx))) {
				//fmt.Println("int", int(C.Z3_stats_get_uint_value(s.RawCtx, s.Stats, C.uint(idx))))
				return int(C.Z3_stats_get_uint_value(s.RawCtx, s.Stats, C.uint(idx)))
			} else {
				//fmt.Println("double", int(C.Z3_stats_get_double_value(s.RawCtx, s.Stats, C.uint(idx))))
				return int(C.Z3_stats_get_double_value(s.RawCtx, s.Stats, C.uint(idx)))
			}
		}
	}
	return 0
}

func (s *Statistics) Length() int {
	return int(C.Z3_stats_size(s.RawCtx, s.Stats))
}
