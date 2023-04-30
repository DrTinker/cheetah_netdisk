package start

import (
	"NetDesk/client"
	"NetDesk/infrastructure/db"
)

func InitDB() {
	driver, source, err := client.GetConfigClient().GetDBConfig()
	if err != nil {
		panic(err)
	}
	impl, err := db.NewDBClientImpl(driver, source)
	if err != nil {
		panic(err)
	}
	client.InitDBClient(impl)
}
