package slice

func InStr(find string, arr []string) bool {
	for _, value := range arr {
		if value == find {
			return true
		}
	}
	return false
}

func InInt(find int, arr []int) bool {
	for _, value := range arr {
		if value == find {
			return true
		}
	}
	return false
}
