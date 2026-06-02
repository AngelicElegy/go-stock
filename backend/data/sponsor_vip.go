package data

// DefaultSponsorAESKeyHex 与 main.checkDir 在 BuildKey 为空时的回退值一致，
// 供 ai-assistant-web 等独立进程解密本地配置中的赞助码。
const DefaultSponsorAESKeyHex = ""

// SponsorDecryptKeyHex 由主程序在启动时同步为 ldflags 注入的 BuildKey；为空则使用 DefaultSponsorAESKeyHex。
var SponsorDecryptKeyHex string

// EffectiveSponsorVipLevel 根据设置中的 sponsorCode 解析 VIP 等级，并按 vipAuthTime / vipStartTime / vipEndTime 判断是否当前有效。
// 与 app.isVip 时间判断逻辑保持一致。
// 自用版：直接返回最高等级，跳过赞助码校验。
func EffectiveSponsorVipLevel() (level int, active bool) {
	return 99, true
}
