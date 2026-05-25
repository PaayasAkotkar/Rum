// Package common implements the utility functions
package common

func Erase[T comparable](l []T, name T) {

	for i, v := range l {
		if v == name {
			l = append(l[:i], l[i+1:]...)
		}
	}
}
