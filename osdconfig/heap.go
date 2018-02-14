package osdconfig

// a heap needed to store jobs
type jobHeap []*dataWrite

func (d jobHeap) Len() int {
	return len(d)
}

// max heap with latest job item at top of the heap
func (d jobHeap) Less(i, j int) bool {
	if d[i].Time > d[j].Time {
		return true
	} else {
		return false
	}
}

func (d jobHeap) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d *jobHeap) Push(x interface{}) {
	*d = append(*d, x.(*dataWrite))
}

func (d *jobHeap) Pop() interface{} {
	x := (*d)[len(*d)-1]
	*d = (*d)[:len(*d)-1]
	return x
}
