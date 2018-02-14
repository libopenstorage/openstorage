package osdconfig

// Whenever a change occurs in kvdb that causes it to execute registered callbacks
// it sends the data in that callback. It is required to push such packets of
// data into a heap that is sorted by the timestamp (latest first)

// a heap is then used to process subsequent client callbacks that work with the data

// a heap needed to store jobs
type jobHeap []*dataWrite

func (d jobHeap) Len() int {
	return len(d)
}

// min heap with latest job item at top of the heap
func (d jobHeap) Less(i, j int) bool {
	if d[i].Time < d[j].Time {
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
