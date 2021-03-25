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
