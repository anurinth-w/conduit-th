package service

var validTransitions = map[string][]string{
"open":            {"assigned", "duplicate"},
"assigned":        {"in_progress", "open"},
"in_progress":     {"done", "pending_surface"},
"pending_surface": {"done"},
"done":            {},
"duplicate":       {},
}

func canTransition(from, to string) bool {
allowed, ok := validTransitions[from]
if !ok {
return false
}
for _, s := range allowed {
if s == to {
return true
}
}
return false
}
