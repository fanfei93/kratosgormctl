package kratosgormctl

import "gorm.io/gorm"

type Config struct {
	db *gorm.DB // db connection

	TableName        string
	EntityOutPath    string
	EntityPkgPath    string
	EntityStructName string
	RepoOutPath      string
	RepoFileName     string
	RepoPkgPath      string
	RepoStructName   string
	BizStructName    string
}
