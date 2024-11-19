package helper

func SliceRemoveFast[T any](slice []T, index int) []T {
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func SliceRemove[T comparable](slice []T, element T) []T {
    j := 0
    for i := 0; i < len(slice); i++ {
        if slice[i] != element {
            slice[j] = slice[i]
            j++
        }
    }
    return slice[:j]
}
