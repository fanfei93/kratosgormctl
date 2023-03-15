package kratosgormctl

import (
	"bytes"
	"context"
	"fmt"
	gentemplate "github.com/fanfei93/kratosgormctl/template"
	"gorm.io/gen"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Generator struct {
	Config
}

func NewGenerator(cfg Config) *Generator {
	generator := &Generator{
		Config: cfg,
	}

	return generator
}

func (g *Generator) UseDB(db *gorm.DB) {
	if db != nil {
		g.db = db
	}
}

// info logger
func (g *Generator) info(logInfos ...string) {
	for _, l := range logInfos {
		g.db.Logger.Info(context.Background(), l)
		log.Println(l)
	}
}

func (g *Generator) Execute() {
	g.info("Start generating code.")
	g.generateEntityFile()
	g.generateRepoFile()
	g.info("Generate code done.")
}

func (g *Generator) generateEntityFile() {
	generator := gen.NewGenerator(gen.Config{
		ModelPkgPath: g.EntityOutPath,
	})

	generator.UseDB(g.db)

	if g.EntityStructName != "" {
		generator.GenerateModelAs(g.TableName, g.EntityStructName)
	} else {
		generator.GenerateModel(g.TableName)
		g.EntityStructName = getCamelName(g.TableName)
	}
	generator.Execute()
}

func (g *Generator) generateRepoFile() {
	structName := getCamelName(g.TableName)
	if g.RepoStructName != "" {
		structName = g.RepoStructName
	}
	outerInterfaceName := structName + "Repo"
	innerInterfaceName := strings.ToLower(outerInterfaceName[:1]) + outerInterfaceName[1:]
	defaultModelName := "default" + outerInterfaceName

	data := &gentemplate.GenBaseStruct{
		EntityPackageName:  g.EntityPkgPath,
		RepoPackageName:    g.RepoPkgPath,
		InnerInterfaceName: innerInterfaceName,
		OuterInterfaceName: outerInterfaceName,
		DefaultModelName:   defaultModelName,
		BizStructName:      g.BizStructName,
		BizRepoName:        g.BizRepoName,
		EntityStructName:   g.EntityStructName,
		TableName:          g.TableName,
	}

	if err := os.MkdirAll(g.RepoOutPath, os.ModePerm); err != nil {
		err := fmt.Errorf("make dir outpath(%s) fail: %s", g.RepoOutPath, err)
		panic(err)
	}

	g.fillGenEntityFile(data)
	g.fillGenRepoFile(data)
	g.fillCustomRepoFile(data)

}

func (g *Generator) fillGenEntityFile(data *gentemplate.GenBaseStruct) {
	content, err := g.getBaseEntityContent(data)
	if err != nil {
		panic(err)
	}
	baseModelPath := g.EntityOutPath + "/" + g.TableName + ".go"
	if isExist(baseModelPath) {
		return
	}
	err = os.WriteFile(baseModelPath, content, 0666)
	if err != nil {
		panic(err)
	}
	exec.Command("goimports", "-l", "-w", baseModelPath).Output()
	exec.Command("gofmt", "-l", "-w", baseModelPath).Output()

	g.info(baseModelPath + " Done")
}

func (g *Generator) fillGenRepoFile(data *gentemplate.GenBaseStruct) {
	content, err := g.getBaseRepoContent(data)
	if err != nil {
		panic(err)
	}
	baseModelPath := g.RepoOutPath + "/" + g.RepoFileName + "_model_gen.go"
	err = os.WriteFile(baseModelPath, content, 0666)
	if err != nil {
		panic(err)
	}
	exec.Command("goimports", "-l", "-w", baseModelPath).Output()
	exec.Command("gofmt", "-l", "-w", baseModelPath).Output()

	g.info(baseModelPath + " Done")
}

func (g *Generator) fillCustomRepoFile(data *gentemplate.GenBaseStruct) {
	content, err := g.getCustomRepoContent(data)
	if err != nil {
		panic(err)
	}
	baseModelPath := g.RepoOutPath + "/" + g.RepoFileName + "_model.go"
	if isExist(baseModelPath) {
		return
	}
	err = os.WriteFile(baseModelPath, content, 0666)
	if err != nil {
		panic(err)
	}
	exec.Command("goimports", "-l", "-w", baseModelPath).Output()
	exec.Command("gofmt", "-l", "-w", baseModelPath).Output()

	g.info(baseModelPath + " Done")
}

func (g *Generator) getBaseEntityContent(data *gentemplate.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_entity_base").Parse(gentemplate.GetGenEntityCustomTemplate())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = parse.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) getBaseRepoContent(data *gentemplate.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_base").Parse(gentemplate.GetGenRepoBaseTemplate())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = parse.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) getCustomRepoContent(data *gentemplate.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_custom").Parse(gentemplate.GetGenRepoCustomTemplate())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = parse.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
