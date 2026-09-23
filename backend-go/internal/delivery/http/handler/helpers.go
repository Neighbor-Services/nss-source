package handler

import (
	"strings"

	"github.com/google/uuid"
)

// cleanUUID parses a string into a uuid.UUID, gracefully stripping any known entity prefixes
// (e.g., fav_, usr_, pro_, req_, apt_, rev_, msg_) and trailing slashes.
func cleanUUID(s string) (uuid.UUID, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/")
	if idx := strings.Index(s, "_"); idx != -1 && idx < len(s)-1 {
		candidate := s[idx+1:]
		if u, err := uuid.Parse(candidate); err == nil {
			return u, nil
		}
	}
	return uuid.Parse(s)
}
