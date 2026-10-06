package calculator

import (
	"errors"
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		op       string
		operands []float64
		unit     AngleUnit
		want     float64
	}{
		{"add", []float64{2, 3}, Radians, 5},
		{"subtract", []float64{2, 3}, Radians, -1},
		{"multiply", []float64{4, 2.5}, Radians, 10},
		{"divide", []float64{7, 2}, Radians, 3.5},
		{"mod", []float64{7, 3}, Radians, 1},
		{"power", []float64{2, 10}, Radians, 1024},
		{"root", []float64{27, 3}, Radians, 3},
		{"root", []float64{-8, 3}, Radians, -2},
		{"log", []float64{8, 2}, Radians, 3},
		{"sqrt", []float64{16}, Radians, 4},
		{"cbrt", []float64{-27}, Radians, -3},
		{"abs", []float64{-4}, Radians, 4},
		{"reciprocal", []float64{4}, Radians, 0.25},
		{"factorial", []float64{5}, Radians, 120},
		{"factorial", []float64{0}, Radians, 1},
		{"exp", []float64{0}, Radians, 1},
		{"ln", []float64{math.E}, Radians, 1},
		{"log10", []float64{1000}, Radians, 3},
		{"log2", []float64{8}, Radians, 3},
		{"sin", []float64{math.Pi / 2}, Radians, 1},
		{"sin", []float64{30}, Degrees, 0.5},
		{"sin", []float64{180}, Degrees, 0},
		{"cos", []float64{-90}, Degrees, 0},
		{"cos", []float64{360}, Degrees, 1},
		{"tan", []float64{45}, Degrees, 1},
		{"asin", []float64{1}, Degrees, 90},
		{"acos", []float64{1}, Radians, 0},
		{"atan", []float64{1}, Degrees, 45},
		{"tanh", []float64{0}, Radians, 0},
	}
	for _, tt := range tests {
		got, err := Evaluate(tt.op, tt.operands, tt.unit)
		if err != nil {
			t.Errorf("%s%v [%s]: unexpected error: %v", tt.op, tt.operands, tt.unit, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("%s%v [%s] = %v, want %v", tt.op, tt.operands, tt.unit, got, tt.want)
		}
	}
}

func TestEvaluateErrors(t *testing.T) {
	tests := []struct {
		op       string
		operands []float64
		unit     AngleUnit
		want     error
	}{
		{"nope", []float64{1}, Radians, ErrUnknownOperation},
		{"add", []float64{1}, Radians, ErrOperandCount},
		{"sqrt", nil, Radians, ErrOperandCount},
		{"divide", []float64{1, 0}, Radians, ErrDivisionByZero},
		{"mod", []float64{1, 0}, Radians, ErrDivisionByZero},
		{"power", []float64{0, -1}, Radians, ErrDivisionByZero},
		{"reciprocal", []float64{0}, Radians, ErrDivisionByZero},
		{"sqrt", []float64{-1}, Radians, ErrDomain},
		{"root", []float64{-16, 2}, Radians, ErrDomain},
		{"root", []float64{8, 0}, Radians, ErrDomain},
		{"ln", []float64{0}, Radians, ErrDomain},
		{"log", []float64{8, 1}, Radians, ErrDomain},
		{"factorial", []float64{2.5}, Radians, ErrDomain},
		{"factorial", []float64{-1}, Radians, ErrDomain},
		{"factorial", []float64{171}, Radians, ErrOverflow},
		{"asin", []float64{2}, Radians, ErrDomain},
		{"tan", []float64{90}, Degrees, ErrDomain},
		{"tan", []float64{-270}, Degrees, ErrDomain},
		{"exp", []float64{1000}, Radians, ErrOverflow},
		{"power", []float64{10, 400}, Radians, ErrOverflow},
	}
	for _, tt := range tests {
		_, err := Evaluate(tt.op, tt.operands, tt.unit)
		if !errors.Is(err, tt.want) {
			t.Errorf("%s%v [%s]: error = %v, want %v", tt.op, tt.operands, tt.unit, err, tt.want)
		}
	}
}

func TestParseAngleUnit(t *testing.T) {
	for in, want := range map[string]AngleUnit{"": Radians, "rad": Radians, "deg": Degrees} {
		if got, err := ParseAngleUnit(in); err != nil || got != want {
			t.Errorf("ParseAngleUnit(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := ParseAngleUnit("grad"); !errors.Is(err, ErrInvalidAngleUnit) {
		t.Errorf("ParseAngleUnit(\"grad\") error = %v, want %v", err, ErrInvalidAngleUnit)
	}
}

func TestOperationsSorted(t *testing.T) {
	ops := Operations()
	if len(ops) != len(operations) {
		t.Fatalf("Operations() returned %d, want %d", len(ops), len(operations))
	}
	for i := 1; i < len(ops); i++ {
		if ops[i-1].Name >= ops[i].Name {
			t.Fatalf("not sorted: %q before %q", ops[i-1].Name, ops[i].Name)
		}
	}
}
