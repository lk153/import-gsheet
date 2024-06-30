package configs

type DbEnv struct {
	DbUser           string `envName:"DB_USER"`
	DbPassword       string `envName:"DB_PASSWORD"`
	DbHost           string `envName:"DB_HOST"`
	DbName           string `envName:"DB_NAME"`
	DbMaxConnections int    `envName:"NV_DB_MAX_CONNS" defaultValue:"10"`
}
