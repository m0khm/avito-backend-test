package errors

import "testing"

func TestAppError_Error(t *testing.T) {
	err := New("CODE", "message", 400)

	if err.Error() != "message" {
		t.Fatalf("expected message, got %s", err.Error())
	}
}

func TestIs(t *testing.T) {
	err := New(ErrForbidden.Code, "forbidden", 403)

	if !Is(err, ErrForbidden) {
		t.Fatal("expected Is(err, ErrForbidden) == true")
	}
	if Is(err, ErrUnauthorized) {
		t.Fatal("expected Is(err, ErrUnauthorized) == false")
	}
}