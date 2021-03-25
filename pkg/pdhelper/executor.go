package pdhelper

const evictLeaderScheduler = "evict-leader-scheduler"

type Executor interface {
	AddEvictScheduler(storeId uint) error
	RemoveEvictScheduler(storeId uint) error
	ListStores() ([]Store, error)
	ListEvictedStore() ([]Store, error)
}
