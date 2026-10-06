// Package calculator implements the scientific operations exposed by the API.
package calculator

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var (
	ErrUnknownOperation = errors.New("unknown operation")
	ErrOperandCount     = errors.New("wrong number of operands")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrDomain           = errors.New("operand outside the operation's domain")
	ErrOverflow         = errors.New("result is too large to represent")
	ErrInvalidAngleUnit = errors.New("invalid angle unit")
)

// AngleUnit is the unit used by trigonometric operations for angles.
type AngleUnit string

const (
	Radians AngleUnit = "rad"
	Degrees AngleUnit = "deg"
)

// ParseAngleUnit accepts "rad" or "deg"; an empty string means radians.
func ParseAngleUnit(s string) (AngleUnit, error) {
	switch AngleUnit(s) {
	case "", Radians:
		return Radians, nil
	case Degrees:
		return Degrees, nil
	}
	return "", fmt.Errorf("%w: %q (use \"rad\" or \"deg\")", ErrInvalidAngleUnit, s)
}

// Operation describes one operation the calculator can evaluate.
type Operation struct {
	Name          string `json:"name"`
	Arity         int    `json:"arity"`
	Description   string `json:"description"`
	UsesAngleUnit bool   `json:"uses_angle_unit"`

	fn func(x []float64, unit AngleUnit) (float64, error)
}

var operations = index([]Operation{
	binary("add", "x + y", func(x, y float64) (float64, error) { return x + y, nil }),
	binary("subtract", "x - y", func(x, y float64) (float64, error) { return x - y, nil }),
	binary("multiply", "x * y", func(x, y float64) (float64, error) { return x * y, nil }),
	binary("divide", "x / y", divide),
	binary("mod", "Remainder of x / y", mod),
	binary("power", "x raised to the power y", power),
	binary("root", "y-th root of x", root),
	binary("log", "Logarithm of x in base y", logBase),

	unary("sqrt", "Square root of x", sqrt),
	unary("cbrt", "Cube root of x", plain(math.Cbrt)),
	unary("abs", "Absolute value of x", plain(math.Abs)),
	unary("reciprocal", "1 / x", reciprocal),
	unary("factorial", "x! for a non-negative integer x", factorial),
	unary("exp", "e raised to the power x", plain(math.Exp)),
	unary("ln", "Natural logarithm of x", positive("ln", math.Log)),
	unary("log10", "Base-10 logarithm of x", positive("log10", math.Log10)),
	unary("log2", "Base-2 logarithm of x", positive("log2", math.Log2)),

	angleIn("sin", "Sine of the angle x", sin),
	angleIn("cos", "Cosine of the angle x", cos),
	angleIn("tan", "Tangent of the angle x", tan),
	angleOut("asin", "Arcsine of x, as an angle", math.Asin),
	angleOut("acos", "Arccosine of x, as an angle", math.Acos),
	angleOut("atan", "Arctangent of x, as an angle", math.Atan),
	unary("sinh", "Hyperbolic sine of x", plain(math.Sinh)),
	unary("cosh", "Hyperbolic cosine of x", plain(math.Cosh)),
	unary("tanh", "Hyperbolic tangent of x", plain(math.Tanh)),
})

// Operations returns every supported operation, sorted by name.
func Operations() []Operation {
	ops := make([]Operation, 0, len(operations))
	for _, op := range operations {
		ops = append(ops, op)
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
	return ops
}

// Lookup returns the operation with the given name.
func Lookup(name string) (Operation, bool) {
	op, ok := operations[name]
	return op, ok
}

// Evaluate applies the named operation to the operands.
func Evaluate(name string, operands []float64, unit AngleUnit) (float64, error) {
	op, ok := operations[name]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownOperation, name)
	}
	if len(operands) != op.Arity {
		return 0, fmt.Errorf("%w: %s takes %d, got %d", ErrOperandCount, name, op.Arity, len(operands))
	}
	r, err := op.fn(operands, unit)
	if err != nil {
		return 0, err
	}
	switch {
	case math.IsNaN(r):
		return 0, fmt.Errorf("%w: %s%v", ErrDomain, name, operands)
	case math.IsInf(r, 0):
		return 0, fmt.Errorf("%w: %s%v", ErrOverflow, name, operands)
	}
	return r, nil
}

func index(ops []Operation) map[string]Operation {
	m := make(map[string]Operation, len(ops))
	for _, op := range ops {
		m[op.Name] = op
	}
	return m
}

func unary(name, desc string, f func(x float64) (float64, error)) Operation {
	return Operation{Name: name, Arity: 1, Description: desc,
		fn: func(a []float64, _ AngleUnit) (float64, error) { return f(a[0]) }}
}

func binary(name, desc string, f func(x, y float64) (float64, error)) Operation {
	return Operation{Name: name, Arity: 2, Description: desc,
		fn: func(a []float64, _ AngleUnit) (float64, error) { return f(a[0], a[1]) }}
}

// angleIn builds an operation whose operand is an angle in the requested unit.
func angleIn(name, desc string, f func(x float64, unit AngleUnit) (float64, error)) Operation {
	return Operation{Name: name, Arity: 1, Description: desc, UsesAngleUnit: true,
		fn: func(a []float64, unit AngleUnit) (float64, error) { return f(a[0], unit) }}
}

// angleOut builds an operation whose result is an angle in the requested unit.
func angleOut(name, desc string, f func(float64) float64) Operation {
	return Operation{Name: name, Arity: 1, Description: desc, UsesAngleUnit: true,
		fn: func(a []float64, unit AngleUnit) (float64, error) {
			r := f(a[0])
			if unit == Degrees {
				r = r * 180 / math.Pi
			}
			return r, nil
		}}
}

func plain(f func(float64) float64) func(float64) (float64, error) {
	return func(x float64) (float64, error) { return f(x), nil }
}

func positive(name string, f func(float64) float64) func(float64) (float64, error) {
	return func(x float64) (float64, error) {
		if x <= 0 {
			return 0, fmt.Errorf("%w: %s needs x > 0", ErrDomain, name)
		}
		return f(x), nil
	}
}

func divide(x, y float64) (float64, error) {
	if y == 0 {
		return 0, ErrDivisionByZero
	}
	return x / y, nil
}

func mod(x, y float64) (float64, error) {
	if y == 0 {
		return 0, ErrDivisionByZero
	}
	return math.Mod(x, y), nil
}

func power(x, y float64) (float64, error) {
	if x == 0 && y < 0 {
		return 0, ErrDivisionByZero
	}
	return math.Pow(x, y), nil
}

func root(x, n float64) (float64, error) {
	if n == 0 {
		return 0, fmt.Errorf("%w: root index must not be 0", ErrDomain)
	}
	if x < 0 {
		if n != math.Trunc(n) || math.Mod(n, 2) == 0 {
			return 0, fmt.Errorf("%w: negative x needs an odd integer root index", ErrDomain)
		}
		r, err := power(-x, 1/n)
		return -r, err
	}
	return power(x, 1/n)
}

func logBase(x, base float64) (float64, error) {
	if x <= 0 || base <= 0 || base == 1 {
		return 0, fmt.Errorf("%w: log needs x > 0 and a positive base other than 1", ErrDomain)
	}
	return math.Log(x) / math.Log(base), nil
}

func sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, fmt.Errorf("%w: sqrt needs x >= 0", ErrDomain)
	}
	return math.Sqrt(x), nil
}

func reciprocal(x float64) (float64, error) {
	return divide(1, x)
}

func factorial(x float64) (float64, error) {
	if x < 0 || x != math.Trunc(x) {
		return 0, fmt.Errorf("%w: factorial needs a non-negative integer", ErrDomain)
	}
	if x > 170 {
		return 0, fmt.Errorf("%w: %v!", ErrOverflow, x)
	}
	r := 1.0
	for i := 2.0; i <= x; i++ {
		r *= i
	}
	return r, nil
}

// rightAngle reports which multiple of 90° (mod 360) deg is, so sin/cos/tan
// return exact values there instead of tiny floating-point residues.
func rightAngle(deg float64) (quadrant int, ok bool) {
	if math.Mod(deg, 90) != 0 {
		return 0, false
	}
	q := math.Mod(deg/90, 4)
	if q < 0 {
		q += 4
	}
	return int(q), true
}

func sin(x float64, unit AngleUnit) (float64, error) {
	if unit == Degrees {
		if q, ok := rightAngle(x); ok {
			return [4]float64{0, 1, 0, -1}[q], nil
		}
		x = x * math.Pi / 180
	}
	return math.Sin(x), nil
}

func cos(x float64, unit AngleUnit) (float64, error) {
	if unit == Degrees {
		if q, ok := rightAngle(x); ok {
			return [4]float64{1, 0, -1, 0}[q], nil
		}
		x = x * math.Pi / 180
	}
	return math.Cos(x), nil
}

func tan(x float64, unit AngleUnit) (float64, error) {
	if unit == Degrees {
		if q, ok := rightAngle(x); ok {
			if q%2 == 1 {
				return 0, fmt.Errorf("%w: tan is undefined at %v°", ErrDomain, x)
			}
			return 0, nil
		}
		x = x * math.Pi / 180
	}
	return math.Tan(x), nil
}
