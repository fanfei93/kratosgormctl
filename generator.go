package kratosgormctl

import (
	"bytes"
	"context"
	"fmt"
	tmpl "github.com/fanfei93/kratosgormctl/template"
	"gorm.io/gen"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	// tmpl "gorm.io/gen/internal/template"
)

type Generator struct {
	Config
}

func NewGenerator(cfg Config) *Generator {
	if cfg.EntityFileName == "" {
		cfg.EntityFileName = cfg.TableName
	}

	if cfg.EntityOutPath != cfg.RepoOutPath {
		cfg.separateEntity = true
	}

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
	g.generateFile()
	g.info("Generate code done.")
}

func (g *Generator) getEntityGenContent() []byte {
	generator := gen.NewGenerator(gen.Config{
		ModelPkgPath: g.EntityOutPath,
	})

	generator.UseDB(g.db)

	data := generator.GenerateModelAs(g.TableName, g.EntityStructName)
	if g.EntityStructName == "" {
		data = generator.GenerateModel(g.TableName)
		g.EntityStructName = getCamelName(g.TableName)
	}

	if data == nil || !data.Generated {
		panic("getEntityFileContent failed with data invalid")
	}

	var buf bytes.Buffer
	err := render(tmpl.Model, &buf, data)
	if err != nil {
		panic(err)
	}

	for _, method := range data.ModelMethods {
		err = render(tmpl.ModelMethod, &buf, method)
		if err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func (g *Generator) getEntityGenContentWithSeparate() []byte {
	generator := gen.NewGenerator(gen.Config{
		ModelPkgPath: g.EntityOutPath,
	})

	generator.UseDB(g.db)

	data := generator.GenerateModelAs(g.TableName, g.EntityStructName)
	if g.EntityStructName == "" {
		data = generator.GenerateModel(g.TableName)
		g.EntityStructName = getCamelName(g.TableName)
	}

	if data == nil || !data.Generated {
		panic("getEntityFileContent failed with data invalid")
	}

	var buf bytes.Buffer
	err := render(tmpl.ModelWithSeparate, &buf, data)
	if err != nil {
		panic(err)
	}

	for _, method := range data.ModelMethods {
		err = render(tmpl.ModelMethod, &buf, method)
		if err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func render(tmpl string, wr io.Writer, data interface{}) error {
	t, err := template.New(tmpl).Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(wr, data)
}

func (g *Generator) generateFile() {
	structName := getCamelName(g.TableName)
	if g.RepoStructName != "" {
		structName = g.RepoStructName
	}
	outerInterfaceName := structName + "Repo"
	innerInterfaceName := strings.ToLower(outerInterfaceName[:1]) + outerInterfaceName[1:]
	defaultModelName := "default" + outerInterfaceName

	data := &tmpl.GenBaseStruct{
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

	if g.separateEntity {
		genContent := g.getEntityGenContentWithSeparate()
		customContent, err := g.getEntityCustomContentWithSeparate(data)
		if err != nil {
			panic(err)
		}

		modelFile := g.EntityOutPath + "/" + g.EntityFileName + ".gen.go"
		g.fillGenEntityFile(genContent, modelFile)
		g.fillCustomEntityFile(customContent, data)
		g.fillGenRepoFile(data, nil)
		g.fillCustomRepoFile(data, true)
	} else {
		genContent := g.getEntityGenContent()
		g.fillGenRepoFile(data, genContent)
		g.fillCustomRepoFile(data, false)
	}
}

func (g *Generator) fillGenEntityFile(content []byte, baseModelPath string) {
	err := os.WriteFile(baseModelPath, content, 0666)
	if err != nil {
		panic(err)
	}
	exec.Command("goimports", "-l", "-w", baseModelPath).Output()
	exec.Command("gofmt", "-l", "-w", baseModelPath).Output()

	g.info(baseModelPath + " Done")
}

func (g *Generator) fillCustomEntityFile(content []byte, data *tmpl.GenBaseStruct) {
	baseModelPath := g.EntityOutPath + "/" + g.TableName + ".go"
	if isExist(baseModelPath) {
		return
	}
	err := os.WriteFile(baseModelPath, content, 0666)
	if err != nil {
		panic(err)
	}
	exec.Command("goimports", "-l", "-w", baseModelPath).Output()
	exec.Command("gofmt", "-l", "-w", baseModelPath).Output()

	g.info(baseModelPath + " Done")
}

func (g *Generator) fillGenRepoFile(data *tmpl.GenBaseStruct, entityContent []byte) {
	var err error
	var content []byte
	if len(entityContent) > 0 {
		data.EntityContent = string(entityContent)
		content, err = g.getBaseRepoWithEntityContent(data)
	} else {
		content, err = g.getBaseRepoContent(data)
		if err != nil {
			panic(err)
		}
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

func (g *Generator) fillCustomRepoFile(data *tmpl.GenBaseStruct, separateEntity bool) {
	var content []byte
	var err error
	if separateEntity {
		content, err = g.getCustomRepoContent(data)
		if err != nil {
			panic(err)
		}
	} else {
		content, err = g.getCustomRepoWithEntityContent(data)
		if err != nil {
			panic(err)
		}
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

func (g *Generator) getEntityCustomContentWithSeparate(data *tmpl.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_entity_base").Parse(tmpl.GetGenEntityCustomTemplate())
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

func (g *Generator) getBaseRepoContent(data *tmpl.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_base").Parse(tmpl.GetGenRepoBaseTemplate())
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
func (g *Generator) getBaseRepoWithEntityContent(data *tmpl.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_base").Parse(tmpl.GetGenRepoBaseWithEntityTemplate())
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

func (g *Generator) getCustomRepoContent(data *tmpl.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_custom").Parse(tmpl.GetGenRepoCustomTemplate())
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

func (g *Generator) getCustomRepoWithEntityContent(data *tmpl.GenBaseStruct) ([]byte, error) {
	parse, err := template.New("gen_repo_custom_with_entity").Parse(tmpl.GetGenRepoWithEntityCustomTemplate())
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
