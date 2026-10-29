package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBoundedReturnsResult(t *testing.T) {
	v, err := bounded(context.Background(), time.Second, func(context.Context) (string, error) {
		return "DB1", nil
	})
	if err != nil || v != "DB1" {
		t.Errorf("bounded() = %q, %v, want DB1, nil", v, err)
	}
}

func TestBoundedGivesUpOnCallIgnoringContext(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	start := time.Now()
	_, err := bounded(context.Background(), 50*time.Millisecond, func(context.Context) (struct{}, error) {
		<-release // a driver call that ignores the cancelled context
		return struct{}{}, nil
	})
	if err == nil {
		t.Fatal("bounded() error = nil, want an error for a call that never returns")
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond+startupGrace+time.Second {
		t.Errorf("bounded() took %v, want it to give up after the timeout plus the grace", elapsed)
	}
}

func TestBoundedStopsWhenContextIsDone(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := bounded(ctx, time.Minute, func(context.Context) (struct{}, error) {
		<-release
		return struct{}{}, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("bounded() error = %v, want context.Canceled", err)
	}
}
