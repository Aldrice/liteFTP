package config

// InitCfg 启动配置
var InitCfg struct {
	DBCfg   dbCfg   `json:"db_cfg"`
	PortCfg portCfg `json:"port_cfg"`
	SrvCfg  srvCfg  `json:"srv_cfg"`
}

// 数据库配置
type dbCfg struct {
	DriverName string `json:"driver_name"`
	CacheMode  string `json:"cache_mode"`
	ForeignKey int    `json:"foreign_key"`
}

// 端口配置
type portCfg struct {
	MinPasvPort int `json:"min_pasv_port"`
	MaxPasvPort int `json:"max_pasv_port"`
	DataPort    int `json:"data_port"`
	LinkPort    int `json:"link_port"`
}

// 服务配置
type srvCfg struct {
	RootDir string `json:"root_dir"`

	EnableAnonymous bool `json:"enable_anonymous"`
	EnablePassive   bool `json:"enable_passive"`
}
