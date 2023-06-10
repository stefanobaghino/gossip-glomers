package main

import "testing"

func expect(t *testing.T, result string, expected string) {
	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
}

func TestIdGen(t *testing.T) {
	genId := idGen(func() string { return "hello" })
	expect(t, genId(), "hello-0")
	expect(t, genId(), "hello-1")
}
