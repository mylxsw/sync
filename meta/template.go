package meta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/buger/jsonparser"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/coll"
	"github.com/mylxsw/go-toolkit/jsonutils"
)

// ParseTemplate 模板解析
func ParseTemplate(templateContent string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"cutoff":         cutOff,
		"implode":        strings.Join,
		"ident":          leftIdent,
		"json":           jsonFormatter,
		"datetime":       datetimeFormat,
		"datetime_noloc": datetimeFormatNoLoc,
		"json_get":       jsonGet,
		"json_gets":      jsonGets,
		"json_flatten":   jsonFlatten,
		"starts_with":    startsWith,
		"ends_with":      endsWith,
		"trim":           strings.Trim,
		"trim_right":     strings.TrimRight,
		"trim_left":      strings.TrimLeft,
		"trim_space":     strings.TrimSpace,
		"format":         fmt.Sprintf,
		"integer":        toInteger,
	}
	var buffer bytes.Buffer
	if err := template.Must(template.New("").Funcs(funcMap).Parse(templateContent)).Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// cutOff 字符串截断
func cutOff(maxLen int, val string) string {
	valRune := []rune(strings.Trim(val, " \n"))

	valLen := len(valRune)
	if valLen <= maxLen {
		return string(valRune)
	}

	return string(valRune[0:maxLen])
}

// 字符串多行缩进
func leftIdent(ident string, message string) string {
	result := coll.MustNew(strings.Split(message, "\n")).Map(func(line string) string {
		return ident + line
	}).Reduce(func(carry string, line string) string {
		return fmt.Sprintf("%s\n%s", carry, line)
	}, "").(string)

	return strings.Trim(result, "\n")
}

// json格式化输出
func jsonFormatter(content string) string {
	var output bytes.Buffer
	if err := json.Indent(&output, []byte(content), "", "    "); err != nil {
		return content
	}

	return output.String()
}

// datetimeFormat 时间格式化，不使用Location
func datetimeFormatNoLoc(datetime time.Time) string {
	return datetime.Format("2006-01-02 15:04:05")
}

// datetimeFormat 时间格式化
func datetimeFormat(datetime time.Time) string {
	loc, _ := time.LoadLocation("Asia/Chongqing")

	return datetime.In(loc).Format("2006-01-02 15:04:05")
}

// jsonFlatten json转换为kv pairs
func jsonFlatten(body string, maxLevel int) []jsonutils.KvPair {
	defer func() {
		if err := recover(); err != nil {
			log.WithFields(log.Fields{
				"func":     "meta.jsonFlatten",
				"body":     body,
				"maxLevel": maxLevel,
			}).Warningf("json解析失败: %s", err)
		}
	}()

	ju, err := jsonutils.New([]byte(body), maxLevel, true)
	if err != nil {
		return make([]jsonutils.KvPair, 0)
	}

	return ju.ToKvPairsArray()
}

// startsWith 判断是字符串开始
func startsWith(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.HasPrefix(haystack, n) {
			return true
		}
	}

	return false
}

// endsWith 判断字符串结尾
func endsWith(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.HasSuffix(haystack, n) {
			return true
		}
	}

	return false
}

// toInteger 转换为整数
func toInteger(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}

	return val
}

// jsonGets 从json中提取单个值，可以使用逗号分割多个key作为备选
func jsonGets(key string, defaultValue string, body string) string {
	keys := strings.Split(key, ",")
	var res string
	for _, k := range keys {
		res = jsonGet(k, "", body)
		if res != "" {
			return res
		}
	}

	return defaultValue
}

// jsonGet 从json中提取单个值
func jsonGet(key string, defaultValue string, body string) string {
	keys := strings.Split(key, ".")

	value, dataType, _, err := jsonparser.Get([]byte(body), keys...)
	if err != nil {
		return defaultValue
	}

	switch dataType {
	case jsonparser.NotExist:
		return defaultValue
	case jsonparser.String:
		if res, err := jsonparser.ParseString(value); err == nil {
			return res
		}
	case jsonparser.Number:
		if res, err := jsonparser.ParseFloat(value); err == nil {
			return strconv.FormatFloat(res, 'f', -1, 32)
		}
		if res, err := jsonparser.ParseInt(value); err == nil {
			return fmt.Sprintf("%d", res)
		}
	case jsonparser.Object:
		fallthrough
	case jsonparser.Array:
		return fmt.Sprintf("%s", value)
	case jsonparser.Boolean:
		if res, err := jsonparser.ParseBoolean(value); err == nil {
			if res {
				return "true"
			} else {
				return "false"
			}
		}
	case jsonparser.Null:
		return "null"
	case jsonparser.Unknown:
		return "unknown"
	}

	return defaultValue
}
