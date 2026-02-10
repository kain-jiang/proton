package version

import (
	"bytes"
	"strconv"
)

// TODO: Add unittest

// 语义版本号
type Version struct {
	components    [3]int
	buildMetadata string
}

// String converts a version back to a string.
func (v *Version) String() string {
	if v == nil {
		return "<nil>"
	}

	var b bytes.Buffer

	for i, c := range v.components {
		if i > 0 {
			b.WriteString(".")
		}
		b.WriteString(strconv.Itoa(c))
	}

	if v.buildMetadata != "" {
		b.WriteString("+")
		b.WriteString(v.buildMetadata)
	}

	return b.String()
}

// Major returns the major release number.
func (v *Version) Major() int { return v.components[0] }

// Minor returns the minor release number.
func (v *Version) Minor() int { return v.components[1] }

// Patch returns the patch release number.
func (v *Version) Patch() int { return v.components[2] }

// BuildMetadata returns the build metadata.
func (v *Version) BuildMetadata() string { return v.buildMetadata }

// Compare returns an integer comparing two versions .
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
func (v *Version) Compare(other *Version) int {
	for i := 0; i < len(v.components) && i < len(other.components); i++ {
		switch {
		case v.components[i] < other.components[i]:
			return -1
		case v.components[i] > other.components[i]:
			return 1
		}
	}

	switch {
	case len(v.components) < len(other.components):
		return -1
	case len(v.components) > len(other.components):
		return 1
	}

	return 0
}

// AtLeast tests if a version is at least equal to a given minimum version.
func (v *Version) AtLeast(min *Version) bool { return v.Compare(min) != -1 }

// WithMajor sets the major for version.
func (v *Version) WithMajor(major int) *Version { v.components[0] = major; return v }

// WithMinor sets the minor for version.
func (v *Version) WithMinor(minor int) *Version { v.components[1] = minor; return v }

// WithPatch sets the patch for version.
func (v *Version) WithPatch(patch int) *Version { v.components[2] = patch; return v }

// WithBuildMetadata sets the build metadata for version.
func (v *Version) WithBuildMetadata(buildMetadata string) *Version {
	v.buildMetadata = buildMetadata
	return v
}
