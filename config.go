package kratosgormctl

import "gorm.io/gorm"

type Config struct {
	db             *gorm.DB // db connection
	separateEntity bool

	TableName        string
	EntityOutPath    string
	EntityPkgPath    string
	EntityFileName   string
	EntityStructName string
	RepoOutPath      string
	RepoFileName     string
	RepoPkgPath      string
	RepoStructName   string
	BizStructName    string
	BizRepoName      string
}
