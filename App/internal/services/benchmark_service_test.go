package services

import "testing"

func TestNormalizeWorkerCountMinimum(t *testing.T) {
	if got := normalizeWorkerCount(0); got != 1 {
		t.Fatalf("esperava 1 worker para entrada 0, recebeu %d", got)
	}
}

func TestNormalizeWorkerCountKeepsCommonValues(t *testing.T) {
	for _, value := range []int{1, 2, 4, 8} {
		got := normalizeWorkerCount(value)
		if got < 1 {
			t.Fatalf("workers inválido para %d: %d", value, got)
		}
	}
}
