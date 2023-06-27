package utils

func Highest(scoreList []float64) float64 {
	maxScore := scoreList[0]
	for i := 0; i < len(scoreList); i++ {
		if scoreList[i] > maxScore {
			maxScore = scoreList[i]
		}
	}
	return maxScore
}

func IndexOf[T comparable](strSlice []T, element T) int {
	for i, str := range strSlice {
		if str == element {
			return i
		}
	}
	return -1
}
