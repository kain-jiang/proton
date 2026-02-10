package executor

import (
	"context"
	"fmt"
	"sync"

	"taskrunner/trait"
)

func cancelJobFunc(cancel func(*trait.Error), jid int) context.CancelFunc {
	return func() {
		cancel(&trait.Error{
			Internal: trait.ECJobCancel,
			Err:      context.Canceled,
			Detail:   fmt.Sprintf("jod %d has been stop", jid),
		})
	}
}

type workNode struct {
	id   int
	next *workNode
}

type freeWorkerQueue struct {
	free *workNode
	// waiting   *workNode
	// signal must never block
	signal          chan int
	locker          sync.Locker
	jobControlCache []*jobControl
	jobControlIndex jobsIndex
}

func newFreeWorkerQueue(parrell int) freeWorkerQueue {
	res := freeWorkerQueue{
		signal:          make(chan int, parrell),
		locker:          &sync.Mutex{},
		jobControlCache: make([]*jobControl, parrell),
		jobControlIndex: newJobsIndex(),
	}
	for i := 0; i < parrell; i++ {
		res.freeNode(i)
	}
	return res
}

func (q *freeWorkerQueue) hole(jc *jobControl) int {
	q.locker.Lock()
	defer q.locker.Unlock()
	if q.free != nil {
		id := q.free.id
		q.free = q.free.next
		q.jobControlCache[id] = jc
		return id
	}
	return -1
}

func (q *freeWorkerQueue) freeNode(id int) {
	q.locker.Lock()
	defer q.locker.Unlock()
	q.free = &workNode{
		id:   id,
		next: q.free,
	}
	jc := q.jobControlCache[id]
	if jc == nil {
		return
	}
	q.jobControlIndex.removeJobControl(jc)
	q.jobControlCache[id] = nil
}

func (q *freeWorkerQueue) push(id int) {
	q.signal <- id
}

func (q *freeWorkerQueue) pop(ctx context.Context) (int, *jobControl) {
	select {
	case index := <-q.signal:
		q.locker.Lock()
		defer q.locker.Unlock()
		j := q.jobControlCache[index]
		return index, j
	case <-ctx.Done():
		return -1, nil
	}
}

type jobControl struct {
	// row locker
	sync.Locker
	cancel  context.CancelFunc
	ctx     context.Context
	job     *trait.JobRecord
	removed bool
}

func (jc *jobControl) SetStatus(s int) {
	jc.Lock()
	defer jc.Unlock()
	jc.job.Target.Status = s
}

func (jc *jobControl) GetStatus() int {
	jc.Lock()
	defer jc.Unlock()
	return jc.job.Target.Status
}

func newJobControl(job *trait.JobRecord) *jobControl {
	return &jobControl{
		Locker: &sync.Mutex{},
		job:    job,
		// cancel: cancel,
		// ctx:    ctx0,
		removed: false,
	}
}

type jobsIndex struct {
	// table locker
	lock sync.Locker
	jobs map[int]*jobControl
}

func newJobsIndex() jobsIndex {
	return jobsIndex{
		lock: &sync.Mutex{},
		jobs: make(map[int]*jobControl),
	}
}

// removeJobControl remove from index without jc's locker.
// lock jc before call this function.
// jc must remove by it's onwer routine/thread.
func (c *jobsIndex) removeJobControl(jc *jobControl) {
	c.lock.Lock()
	defer c.lock.Unlock()
	jcCache, ok := c.jobs[jc.job.ID]
	if jcCache != jc || jcCache == nil || !ok {
		return
	}
	if jc.cancel != nil {
		jc.cancel()
	}
	jc.removed = true

	delete(c.jobs, jc.job.ID)
}

func (c *jobsIndex) IndexOrCreate(id int) *jobControl {
	c.lock.Lock()
	defer c.lock.Unlock()
	jcCache, ok := c.jobs[id]
	if jcCache == nil || !ok {
		job := &trait.JobRecord{ID: id}
		jcCache = newJobControl(job)
		c.jobs[id] = jcCache
	}
	return jcCache
}

func (c *jobsIndex) IndexAndLock(id int) (jc *jobControl) {
	expired := true
	for expired {
		jc = c.IndexOrCreate(id)
		jc.Lock()
		expired = jc.removed
	}
	return
}
