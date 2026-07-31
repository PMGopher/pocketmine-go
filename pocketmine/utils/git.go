package utils

import "strings"

// RepositoryState returns the git hash of the currently checked out HEAD of the given
// repository, or ("", false) on failure. dirty reports whether the repo has local changes.
func RepositoryState(dir string) (hash string, dirty bool, ok bool) {
	out, _, code := ExecuteProcess("git", "-C", dir, "rev-parse", "HEAD")
	out = strings.TrimSpace(out)
	if code != 0 || len(out) != 40 {
		return "", false, false
	}
	if _, _, code := ExecuteProcess("git", "-C", dir, "diff", "--quiet"); code == 1 {
		dirty = true
	} else if _, _, code := ExecuteProcess("git", "-C", dir, "diff", "--cached", "--quiet"); code == 1 {
		dirty = true
	}
	return out, dirty, true
}

// RepositoryStatePretty is infallible: it returns a string representing the git state, or a
// string of zeros on failure, with a "-dirty" suffix if the repo has local changes.
func RepositoryStatePretty(dir string) string {
	hash, dirty, ok := RepositoryState(dir)
	if !ok {
		return strings.Repeat("00", 20)
	}
	if dirty {
		return hash + "-dirty"
	}
	return hash
}
