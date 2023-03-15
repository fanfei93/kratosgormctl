package kratosgormctl

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGenerator_Execute(t *testing.T) {
	db, _ := gorm.Open(mysql.Open("root:password@(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"))
	g := NewGenerator(Config{
		TableName:        "table",
		EntityOutPath:    "./data/entity",
		EntityPkgPath:    "entity",
		EntityStructName: "OssBucket",
		RepoOutPath:      "./data/repo/metadata",
		RepoFileName:     "metadata_oss_bucket",
		RepoPkgPath:      "metadata",
		RepoStructName:   "OssBucketRepo",
		BizStructName:    "biz.OssBucket",
	})

	g.UseDB(db)

	g.Execute()
}
