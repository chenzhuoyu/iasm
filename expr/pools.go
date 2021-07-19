package expr

import (
    `sync`
)

var (
	expressionPool sync.Pool
)

func newExpression() *Expr {
    var v *Expr
    var p interface{}

    /* attemt to grab from pool */
    if p = expressionPool.Get(); p == nil {
        return new(Expr)
    }

    /* clear the expression */
    v = p.(*Expr)
    *v = Expr{}
    return v
}

func freeExpression(p *Expr) {
    expressionPool.Put(p)
}
