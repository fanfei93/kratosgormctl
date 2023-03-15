package kratosgormctl

import (
	"github.com/xxjwxc/public/mybigcamel"
	"os"
	"strings"
)

// getCamelName Big Hump or Capital Letter.大驼峰或者首字母大写
func getCamelName(name string) string {
	return mybigcamel.Marshal(strings.ToLower(name))
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
