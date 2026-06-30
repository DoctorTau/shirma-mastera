package dndsu

import "testing"

func TestParseCatalogFragment(t *testing.T) {
	html := readFixture(t, "catalog_2014_fragment_sample.html")
	entries, err := ParseCatalogFragment(html, "https://dnd.su")
	if err != nil {
		t.Fatalf("ParseCatalogFragment: %v", err)
	}
	if len(entries) != 40 {
		t.Fatalf("len(entries) = %d, want 40", len(entries))
	}

	first := entries[0]
	if first.DndsuID != 7758 {
		t.Errorf("first.DndsuID = %d, want 7758", first.DndsuID)
	}
	if first.NameRu != "Аластра" || first.NameEn != "Alastrah" {
		t.Errorf("first names = %q/%q, want Аластра/Alastrah", first.NameRu, first.NameEn)
	}
	if first.Slug != "7758-alastrah" {
		t.Errorf("first.Slug = %q, want 7758-alastrah", first.Slug)
	}
	if !first.IsUniqueNPC {
		t.Error("first.IsUniqueNPC = false, want true (Аластра is a named NPC)")
	}
	if first.URL != "https://dnd.su/bestiary/7758-alastrah/" {
		t.Errorf("first.URL = %q", first.URL)
	}

	aarakocra := findEntry(entries, 30)
	if aarakocra == nil {
		t.Fatal("entry 30 (aarakocra) not found")
	}
	if aarakocra.CR != "1/4" {
		t.Errorf("aarakocra.CR = %q, want 1/4", aarakocra.CR)
	}
	if aarakocra.IsUniqueNPC {
		t.Error("aarakocra.IsUniqueNPC = true, want false")
	}
}

func findEntry(entries []CatalogEntry, id int) *CatalogEntry {
	for i := range entries {
		if entries[i].DndsuID == id {
			return &entries[i]
		}
	}
	return nil
}

func TestDiffCatalog(t *testing.T) {
	fresh := []CatalogEntry{
		{DndsuID: 1, Slug: "a", NameRu: "Альфа", CR: "1"},
		{DndsuID: 2, Slug: "b", NameRu: "Бета", CR: "2"},
		{DndsuID: 3, Slug: "c", NameRu: "Гамма", CR: "3"},
	}
	previous := map[int]string{
		1: HashCatalogEntry(fresh[0]),                                        // unchanged
		2: HashCatalogEntry(CatalogEntry{Slug: "b", NameRu: "Бета", CR: "1"}), // changed (CR differs)
	}

	diff := DiffCatalog(fresh, previous, HashCatalogEntry)

	if len(diff.New) != 1 || diff.New[0].DndsuID != 3 {
		t.Errorf("New = %+v, want only entry 3", diff.New)
	}
	if len(diff.Changed) != 1 || diff.Changed[0].DndsuID != 2 {
		t.Errorf("Changed = %+v, want only entry 2", diff.Changed)
	}
	if len(diff.Same) != 1 || diff.Same[0].DndsuID != 1 {
		t.Errorf("Same = %+v, want only entry 1", diff.Same)
	}
}
