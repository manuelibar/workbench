package skills

type OverlayRegistry struct{ registries []SkillRegistry }

func NewOverlayRegistry(registries ...SkillRegistry) *OverlayRegistry {
	return &OverlayRegistry{registries: registries}
}

func (r *OverlayRegistry) All() []Bundle { return r.For(true) }

func (r *OverlayRegistry) Get(name string) (Bundle, bool) {
	for _, reg := range r.registries {
		if b, ok := reg.Get(name); ok {
			return b, true
		}
	}
	return Bundle{}, false
}

func (r *OverlayRegistry) For(hasProject bool) []Bundle {
	seen := map[string]bool{}
	out := []Bundle{}
	for _, reg := range r.registries {
		for _, b := range reg.For(hasProject) {
			if seen[b.Name] {
				continue
			}
			seen[b.Name] = true
			out = append(out, b)
		}
	}
	return out
}
