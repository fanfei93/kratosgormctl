# KratosGormctl
目的：为使用kratos框架的项目自动生成基于gorm的data层代码

## 使用方式
### 1. entity和repo分离的场景
```go
    import (
        "github.com/fanfei93/kratosgormctl"
        "gorm.io/driver/mysql"
        "gorm.io/gorm"
    )

    ...

    db, _ := gorm.Open(mysql.Open("root:@(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"))
    g := kratosgormctl.NewGenerator(kratosgormctl.Config{
        EntityOutPath: "./entity",          // entity文件生成位置，推荐使用绝对路径
        EntityPkgPath:    "entity",         // entity文件包名
        TableName:        "user_info",      // 表名
        EntityStructName: "UserEntity",     // 生成的Entity sturct名称
        RepoOutPath:      "./repo/user",    // repo文件生成位置，推荐使用绝对路径
        RepoFileName:     "user_info",      // repo文件名前缀
        RepoPkgPath:      "user",           // repo文件包名
        RepoStructName:   "User",           // repo struct名称
        BizStructName:    "biz.User",       // 依赖的biz层struct名称 package.struct
        BizRepoName:      "biz.UserRepo",   // 依赖的biz层的repo接口名称 packaeg.repo
    })

    g.UseDB(db)

    g.Execute()
```

### 2. entity和repo不分离
将配置项中的EntityOutPath设置为空即可
```go
    import (
        "github.com/fanfei93/kratosgormctl"
        "gorm.io/driver/mysql"
        "gorm.io/gorm"
    )

    ...

    db, _ := gorm.Open(mysql.Open("root:@(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"))
    g := kratosgormctl.NewGenerator(kratosgormctl.Config{
        EntityOutPath:    "",               // entity文件生成位置，推荐使用绝对路径
        TableName:        "user_info",      // 表名
        EntityStructName: "UserEntity",     // 生成的Entity sturct名称
        RepoOutPath:      "./repo/user",    // repo文件生成位置，推荐使用绝对路径
        RepoFileName:     "user_info",      // repo文件名前缀
        RepoPkgPath:      "user",           // repo文件包名
        RepoStructName:   "User",           // repo struct名称
        BizStructName:    "biz.User",       // 依赖的biz层struct名称 package.struct
        BizRepoName:      "biz.UserRepo",   // 依赖的biz层的repo接口名称 packaeg.repo
    })

    g.UseDB(db)

    g.Execute()
```