package domain

type Settings struct {
	IsSysboxEnabled          bool
	IsSysboxWhitelistIgnored bool

	// WhitelistedImages are the user-added repositories allowed to run with elevated access; read-only here,
	// managed through storage.WhitelistStorage.
	WhitelistedImages []string
}

type UpdateSettingsReq struct {
	IsSysboxEnabled          *bool
	IsSysboxWhitelistIgnored *bool
}
