package entities

import (
	"testing"
)

func TestIsTerminalStatus(t *testing.T) {
	if !IsTerminalStatus(OrderStatusDelivered) {
		t.Error("expected true for delivered")
	}
	if !IsTerminalStatus(OrderStatusCancelled) {
		t.Error("expected true for cancelled")
	}
	if IsTerminalStatus(OrderStatusPending) {
		t.Error("expected false for pending")
	}
}

func TestCanTransition(t *testing.T) {
	if !CanTransition(OrderStatusPending, OrderStatusProcessing) {
		t.Error("expected true")
	}
	if !CanTransition(OrderStatusProcessing, OrderStatusShipped) {
		t.Error("expected true")
	}
	if CanTransition(OrderStatusCancelled, OrderStatusPending) {
		t.Error("expected false")
	}
	if CanTransition("invalid", OrderStatusPending) {
		t.Error("expected false")
	}
}

func TestCanTransitionPayment(t *testing.T) {
	if !CanTransitionPayment(PaymentStatusUnpaid, PaymentStatusPaid) {
		t.Error("expected true")
	}
	if CanTransitionPayment(PaymentStatusPaid, PaymentStatusUnpaid) {
		t.Error("expected false")
	}
	if CanTransitionPayment(PaymentStatusRefunded, PaymentStatusPaid) {
		t.Error("expected false")
	}
}
