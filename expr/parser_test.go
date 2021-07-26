package expr

import (
    `testing`

    `github.com/stretchr/testify/assert`
    `github.com/stretchr/testify/require`
)

func TestParser_Eval(t *testing.T) {
    p := new(Parser).SetSource(`3 - 2 * (5 + 6) ** 4 / 7 + (1 << (1234 % 23)) & 0x5436 ^ 0x5a5a - 2 | 1`)
    v, err := p.Parse(nil)
    require.NoError(t, err)
    r, err := v.Evaluate()
    require.NoError(t, err)
    assert.Equal(t, int64(7805), r)
}
