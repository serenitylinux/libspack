package repo

func MockRepo(name string, entries ...Entry) *Repo {
	r := &Repo{Name: name}
	r.entries = make(map[string][]Entry)
	for _, entry := range entries {
		r.addEntry(entry)
	}
	return r
}
