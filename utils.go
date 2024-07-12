package main

import (
	"fmt"
	"github.com/google/uuid"
	"math/big"
)

func atIndexOr[T comparable](index int, or T, array []T) T {
	if len(array) >= index {
		return array[index]
	}
	return or
}

func toBase62(uuid uuid.UUID) string {
	var i big.Int
	i.SetBytes(uuid[:])
	return i.Text(62)
}

func parseBase62(s string) (uuid.UUID, error) {
	var i big.Int
	_, ok := i.SetString(s, 62)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("Failed to parse base62: %s", s)
	}

	var u uuid.UUID
	copy(u[:], i.Bytes())
	return u, nil
}
