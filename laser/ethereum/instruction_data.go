package ethereum

const GAS_SHA3 = 30
const GAS_SHA3WORD = 6

func CalculateSha3Gas(length int) (int, int) {
	gasVal := GAS_SHA3 + GAS_SHA3WORD*(Ceil32(length)/32)
	return gasVal, gasVal
}

// Ceil32 the implementation is in
// https://github.com/ethereum/py-evm/blob/master/eth/_utils/numeric.py
func Ceil32(value int) int {
	remainder := value % 32
	if remainder == 0 {
		return value
	} else {
		return value + 32 - remainder
	}
}
