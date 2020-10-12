package utils

/*
Util for array operations
*/

import(
	"math"
)

func FindMin(array []float64) float64 {

	min := math.MaxFloat64
	for _, element := range array {
		if element < min {
			min = element
		}
	}
	return min
}

func Contains(a []string, x string) bool {
        for _, n := range a {
                if x == n {
                        return true
                }
        }
        return false
}

func AppendMaps(a map[string]string, b map[string]string) (map[string]string){
	temp := make(map[string]string)
	for item, element := range a {
		temp[item] = element
	}
	for item, element := range b {
		temp[item] = element
	}
	return temp
}

func FindMax(a int, b int)(int){
	if(a > b){
		return a
	}
	return b
}

func GetMapKeys(data map[string]map[string]string)([]string){
	keys := []string{}
    for key, _ := range data {
        keys = append(keys, key)
	}
	return keys
}