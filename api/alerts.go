package api

// Implementation of the sort interface for api.Alerts slice

type AlertsSlice []*Alerts

// Len returns the length of slice
func (s AlertsSlice) Len()int {
	return len(s)
}

// Less returns the Alert which has lower Id
func (s AlertsSlice) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

// Swap swaps the position of two alerts in the slice
func (s AlertsSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
