package dndsu

import (
	"crypto/sha256"
	"encoding/hex"
)

// ContentHash hashes raw card HTML so the incremental crawl (§8.3) can tell
// whether a previously-seen monster actually changed before re-parsing and
// re-fetching its cross-linked edition.
func ContentHash(html string) string {
	sum := sha256.Sum256([]byte(html))
	return hex.EncodeToString(sum[:])
}

// CatalogDiff describes what changed between a freshly-fetched catalog and
// what's already stored, keyed by dndsu_id.
type CatalogDiff struct {
	New     []CatalogEntry
	Changed []CatalogEntry
	Same    []CatalogEntry
}

// DiffCatalog compares freshly fetched entries against a map of previously
// seen content hashes (dndsu_id -> content_hash) computed from each entry's
// own fields (since the catalog fragment carries no hash of its own, the
// caller passes a function to hash an entry's comparable fields).
func DiffCatalog(fresh []CatalogEntry, previousHashes map[int]string, hashEntry func(CatalogEntry) string) CatalogDiff {
	var diff CatalogDiff
	for _, e := range fresh {
		hash := hashEntry(e)
		prev, seen := previousHashes[e.DndsuID]
		switch {
		case !seen:
			diff.New = append(diff.New, e)
		case prev != hash:
			diff.Changed = append(diff.Changed, e)
		default:
			diff.Same = append(diff.Same, e)
		}
	}
	return diff
}

// HashCatalogEntry is the default comparable-fields hash used by DiffCatalog.
func HashCatalogEntry(e CatalogEntry) string {
	sum := sha256.Sum256([]byte(e.Slug + "|" + e.NameRu + "|" + e.NameEn + "|" + e.CR))
	return hex.EncodeToString(sum[:])
}
