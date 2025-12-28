package storage

// JoinStrings is a small helper used across packages for building SQL placeholders.
func JoinStrings(strs []string, sep string) string {
	return joinStrings(strs, sep)
}
