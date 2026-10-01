package research

import (
	"regexp"
	"strings"
)

// Path rules per §6: repository-relative, forward slashes, no absolute
// paths, no "..", and a closed character set.
var evidencePathRe = regexp.MustCompile(`^data/[A-Za-z0-9._/-]+$`)

var controlPathRe = regexp.MustCompile(
	`^data/(research-inputs/(capture-inventory\.tsv|build\.json|policies/[A-Za-z0-9._-]+\.json|fetch-logs/[A-Za-z0-9._-]+\.log)|reviews/decisions/[A-Za-z0-9._-]+\.json)$`)

// ValidEvidencePath reports whether p is a well-formed evidence input path.
func ValidEvidencePath(p string) bool {
	if !evidencePathRe.MatchString(p) || strings.Contains(p, "//") {
		return false
	}
	if underStaging(p) {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}

// ValidControlPath reports whether p is a well-formed control-file path.
func ValidControlPath(p string) bool {
	return controlPathRe.MatchString(p)
}

// Control-file path predicates used by the inventory gate (§2.1).
func isControlPath(p string) bool { return controlPathRe.MatchString(p) }

// stagingPrefix is the quarantine lane; no evidence path may live under it.
const stagingPrefix = "data/research-staging/"

func underStaging(p string) bool { return strings.HasPrefix(p, stagingPrefix) }
