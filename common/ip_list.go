package common

// IP 黑白名单功能已随配置文件一并废除。
// 黑名单检测统一返回 false（全部放行）。
func InList(_ string) bool  { return false }
func GetList() []string     { return nil }
func UpDataList(_ bool, _ []string) {}
