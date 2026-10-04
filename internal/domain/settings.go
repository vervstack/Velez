package domain

type Settings struct {
	IsSysboxEnabled          bool
	IsSysboxWhitelistIgnored bool
}

type UpdateSettingsReq struct {
	IsSysboxEnabled          *bool
	IsSysboxWhitelistIgnored *bool
}
