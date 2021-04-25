package types

type ConfigType struct {
	Project       []ProjectType
	RequestParams map[int]string
	Logger        logType
	Redis         RdbType
	RequestMaxNum int
}

type logType struct {
	Filepath string
	Level    int
}

type RdbType struct {
	IsCluster bool
	HostList  []string
	Password  string
}

type ProjectType struct {
	SecretKey string
	Appid     string
	Bizid     string
	Game      string
}
