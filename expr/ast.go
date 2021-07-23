package expr

import (
    `fmt`
)

// Type is tyep expression type.
type Type int

const (
    // CONST indicates that the expression is a constant.
    CONST Type = iota

    // TERM indicates that the expression is a Term reference.
    TERM

    // EXPR indicates that the expression is a unary or binary expression.
    EXPR
)

var typeNames = map[Type]string {
    EXPR  : "Expr",
    TERM  : "Term",
    CONST : "Const",
}

// String returns the string representation of a Type.
func (self Type) String() string {
    if v, ok := typeNames[self]; ok {
        return v
    } else {
        return fmt.Sprintf("expr.Type(%d)", self)
    }
}

// Operator represents an operation to perform when Type is EXPR.
type Operator int

const (
    // ADD performs "Add Expr.Left and Expr.Right".
    ADD Operator = iota

    // SUB performs "Subtract Expr.Left by Expr.Right".
    SUB

    // MUL performs "Multiply Expr.Left by Expr.Right".
    MUL

    // DIV performs "Divide Expr.Left by Expr.Right".
    DIV

    // MOD performs "Modulo Expr.Left by Expr.Right".
    MOD

    // AND performs "Bitwise AND Expr.Left and Expr.Right".
    AND

    // OR performs "Bitwise OR Expr.Left and Expr.Right".
    OR

    // XOR performs "Bitwise XOR Expr.Left and Expr.Right".
    XOR

    // SHL performs "Bitwise Shift Expr.Left to the Left by Expr.Right Bits".
    SHL

    // SHR performs "Bitwise Shift Expr.Left to the Right by Expr.Right Bits".
    SHR

    // NOT performs "Bitwise Invert Expr.Left".
    NOT

    // NEG performs "Negate Expr.Left".
    NEG
)

var operatorNames = map[Operator]string {
    ADD : "Add",
    SUB : "Subtract",
    MUL : "Multiply",
    DIV : "Divide",
    MOD : "Modulo",
    AND : "And",
    OR  : "Or",
    XOR : "ExclusiveOr",
    SHL : "ShiftLeft",
    SHR : "ShiftRight",
    NOT : "Invert",
    NEG : "Negate",
}

// String returns the string representation of a Type.
func (self Operator) String() string {
    if v, ok := operatorNames[self]; ok {
        return v
    } else {
        return fmt.Sprintf("expr.Operator(%d)", self)
    }
}

// Expr represents an expression node.
type Expr struct {
    Type  Type
    Term  Term
    Op    Operator
    Left  *Expr
    Right *Expr
    Const int64
}

// Ref creates an expression from a Term.
func Ref(t Term) (p *Expr) {
    p = newExpression()
    p.Term = t
    p.Type = TERM
    return
}

// Int creates an expression from an integer.
func Int(v int64) (p *Expr) {
    p = newExpression()
    p.Type = CONST
    p.Const = v
    return
}

// Free returns the Expr into pool.
// Any operation performed after Free is undefined behavior.
func (self *Expr) Free() {
    if self != nil {
        self.Term = nil
        self.Left.Free()
        self.Right.Free()
        freeExpression(self)
    }
}

// Evaluate evaluates the expression into an integer.
// It also implements the Term interface.
func (self *Expr) Evaluate() int64 {
    switch self.Type {
        case EXPR  : return self.eval()
        case TERM  : return self.Term.Evaluate()
        case CONST : return self.Const
        default    : panic("invalid expression type: " + self.Type.String())
    }
}

/** Expression Combinator **/

func combine(a *Expr, op Operator, b *Expr) (r *Expr) {
    r = newExpression()
    r.Op = op
    r.Type = EXPR
    r.Left = a
    r.Right = b
    return
}

func (self *Expr) Add(v *Expr) *Expr { return combine(self, ADD, v) }
func (self *Expr) Sub(v *Expr) *Expr { return combine(self, SUB, v) }
func (self *Expr) Mul(v *Expr) *Expr { return combine(self, MUL, v) }
func (self *Expr) Div(v *Expr) *Expr { return combine(self, DIV, v) }
func (self *Expr) Mod(v *Expr) *Expr { return combine(self, MOD, v) }
func (self *Expr) And(v *Expr) *Expr { return combine(self, AND, v) }
func (self *Expr) Or (v *Expr) *Expr { return combine(self, OR , v) }
func (self *Expr) Xor(v *Expr) *Expr { return combine(self, XOR, v) }
func (self *Expr) Shl(v *Expr) *Expr { return combine(self, SHL, v) }
func (self *Expr) Shr(v *Expr) *Expr { return combine(self, SHR, v) }
func (self *Expr) Not()        *Expr { return combine(self, NOT, nil) }
func (self *Expr) Neg()        *Expr { return combine(self, NEG, nil) }

/** Expression Evaluator **/

func (self *Expr) eval() int64 {
    switch self.Op {
        case ADD : return self.Left.Evaluate() + self.Right.Evaluate()
        case SUB : return self.Left.Evaluate() - self.Right.Evaluate()
        case MUL : return self.Left.Evaluate() * self.Right.Evaluate()
        case DIV : return self.Left.Evaluate() / self.Right.Evaluate()
        case MOD : return self.Left.Evaluate() % self.Right.Evaluate()
        case AND : return self.Left.Evaluate() & self.Right.Evaluate()
        case OR  : return self.Left.Evaluate() | self.Right.Evaluate()
        case XOR : return self.Left.Evaluate() ^ self.Right.Evaluate()
        case SHL : return self.Left.Evaluate() << self.Right.Evaluate()
        case SHR : return self.Left.Evaluate() >> self.Right.Evaluate()
        case NOT : return self.unaryNot()
        case NEG : return self.unaryNeg()
        default  : panic("invalid operator: " + self.Op.String())
    }
}

func (self *Expr) unaryNot() int64 {
    if self.Right == nil {
        return ^self.Left.Evaluate()
    } else {
        panic("operator Invert is an unary operator")
    }
}

func (self *Expr) unaryNeg() int64 {
    if self.Right == nil {
        return -self.Left.Evaluate()
    } else {
        panic("operator Negate is an unary operator")
    }
}
