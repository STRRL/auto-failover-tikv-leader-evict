package pdhelper

type PdStore struct {
	Count  int         `json:"count"`
	Stores []StoreItem `json:"stores"`
}

type StoreItem struct {
	Store Store `json:"store"`
}

type Store struct {
	Id      uint   `json:"id"`
	Address string `json:"address"`
}

type PdSchedulerConfig struct {
	StoreIdRanges map[uint][]interface{} `json:"store-id-ranges"`
}

func (it *PdSchedulerConfig) FetchStoreIds() []uint {
	var result []uint
	for k, _ := range it.StoreIdRanges {
		result = append(result, k)
	}
	return result
}
