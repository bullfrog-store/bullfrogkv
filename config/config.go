package config

type Config struct {
	RouteConfig RouteConfig
	StoreConfig StoreConfig
	RaftConfig  RaftConfig
}

type RouteConfig struct {
	Addr string
}

type StoreConfig struct {
	StoreId  uint64
	DataPath string
}

type RaftConfig struct {
	LogGCCountLimit    int
	CompactCheckPeriod int
}
