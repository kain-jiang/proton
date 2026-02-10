package componentmanage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCompare(t *testing.T) {
	r1, err1 := compareComponentVersion("1.4.0", "1.4.0")
	assert.Equal(t, 0, r1, "1.4.0 == 1.4.0")
	assert.Nil(t, err1, "1.4.0 == 1.4.0")

	r2, err2 := compareComponentVersion("1.4.0+1.1.1", "1.4.0+1.1.1")
	assert.Equal(t, 0, r2, "1.4.0+1.1.1 == 1.4.0+1.1.1")
	assert.Nil(t, err2, "1.4.0+1.1.1 == 1.4.0+1.1.1")

	r3, err3 := compareComponentVersion("1.4.0+1.1.0", "1.4.0+1.1.1")
	assert.Equal(t, -1, r3, "1.4.0+1.1.0 < 1.4.0+1.1.1")
	assert.Nil(t, err3, "1.4.0+1.1.0 < 1.4.0+1.1.1")

	r4, err4 := compareComponentVersion("1.4.1", "1.4.0")
	assert.Equal(t, 1, r4, "1.4.1 > 1.4.0")
	assert.Nil(t, err4, "1.4.1 > 1.4.0")
}
