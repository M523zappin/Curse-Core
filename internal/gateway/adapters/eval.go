package adapters

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type EvalAdapter struct {
	profile gateway.ModelProfile
}

func NewEval(profile gateway.ModelProfile) *EvalAdapter {
	return &EvalAdapter{profile: profile}
}

func (a *EvalAdapter) Name() string { return "eval" }
func (a *EvalAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *EvalAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	q := ""
	for _, m := range req.Messages {
		if m.Role == gateway.RoleUser {
			q = m.Content
			break
		}
	}
	expr := extractMathExpr(q)
	if expr == "" {
		return &gateway.Response{
			Message: gateway.Message{Role: gateway.RoleAssistant, Content: "Send a math expression to evaluate, e.g. 2 + 2 or sin(pi/4) * 180 / pi"},
			Done:    true,
		}, nil
	}

	result, err := evalMath(expr)
	if err != nil {
		return &gateway.Response{
			Message: gateway.Message{Role: gateway.RoleAssistant, Content: fmt.Sprintf("Error evaluating %q: %s", expr, err)},
			Done:    true,
		}, nil
	}

	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: fmt.Sprintf("%s = %v", expr, result)},
		Done:    true,
	}, nil
}

func extractMathExpr(q string) string {
	q = strings.TrimSpace(q)
	for _, prefix := range []string{"calc ",=","eval ","math ","compute ","calculate "} {
		q = strings.TrimPrefix(q, prefix)
	}
	q = strings.TrimSpace(q)
	q = strings.Trim(q, `"'`)
	return q
}

func evalMath(expr string) (float64, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseExprFrom(fset, "", expr, 0)
	if err != nil {
		return 0, fmt.Errorf("parse: %w", err)
	}
	return evalNode(parsed)
}

func evalNode(node ast.Node) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT {
			v, _ := strconv.ParseFloat(n.Value, 64)
			return v, nil
		}
		if n.Kind == token.FLOAT {
			v, _ := strconv.ParseFloat(n.Value, 64)
			return v, nil
		}
		return 0, fmt.Errorf("unsupported literal: %s", n.Value)

	case *ast.ParenExpr:
		return evalNode(n.X)

	case *ast.UnaryExpr:
		v, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -v, nil
		}
		return v, nil

	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			return float64(int64(left) % int64(right)), nil
		case token.EQL:
			if left == right { return 1, nil }
			return 0, nil
		case token.NEQ:
			if left != right { return 1, nil }
			return 0, nil
		case token.LSS:
			if left < right { return 1, nil }
			return 0, nil
		case token.GTR:
			if left > right { return 1, nil }
			return 0, nil
		case token.LEQ:
			if left <= right { return 1, nil }
			return 0, nil
		case token.GEQ:
			if left >= right { return 1, nil }
			return 0, nil
		default:
			return 0, fmt.Errorf("unsupported operator: %s", n.Op)
		}

	case *ast.CallExpr:
		fn, ok := n.Fun.(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("unsupported call")
		}
		args := make([]float64, len(n.Args))
		for i, arg := range n.Args {
			v, err := evalNode(arg)
			if err != nil {
				return 0, err
			}
			args[i] = v
		}
		switch fn.Name {
		case "sin":
			return math.Sin(args[0]), nil
		case "cos":
			return math.Cos(args[0]), nil
		case "tan":
			return math.Tan(args[0]), nil
		case "sqrt":
			return math.Sqrt(args[0]), nil
		case "abs":
			return math.Abs(args[0]), nil
		case "pow":
			if len(args) < 2 { return 0, fmt.Errorf("pow needs 2 args") }
			return math.Pow(args[0], args[1]), nil
		case "log":
			return math.Log(args[0]), nil
		case "log10":
			return math.Log10(args[0]), nil
		case "ceil":
			return math.Ceil(args[0]), nil
		case "floor":
			return math.Floor(args[0]), nil
		case "round":
			return math.Round(args[0]), nil
		case "max":
			if len(args) < 2 { return 0, fmt.Errorf("max needs 2 args") }
			return math.Max(args[0], args[1]), nil
		case "min":
			if len(args) < 2 { return 0, fmt.Errorf("min needs 2 args") }
			return math.Min(args[0], args[1]), nil
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		default:
			return 0, fmt.Errorf("unknown function: %s", fn.Name)
		}

	case *ast.Ident:
		switch n.Name {
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		default:
			return 0, fmt.Errorf("unknown identifier: %s", n.Name)
		}

	default:
		return 0, fmt.Errorf("unsupported expression type: %T", node)
	}
}
