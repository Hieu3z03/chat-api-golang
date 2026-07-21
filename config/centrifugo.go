package config

type CentrifugoSettings struct {
	APIURL          string
	APIKey          string
	TokenHMACSecret string
}

func LoadCentrifugoSettings() CentrifugoSettings {
	return CentrifugoSettings{
		APIURL:          getEnvOrDefault("CENTRIFUGO_API_URL", "http://localhost:8000/api"),
		APIKey:          getEnvOrDefault("CENTRIFUGO_API_KEY", "change_me_centrifugo_api_key"),
		TokenHMACSecret: getEnvOrDefault("CENTRIFUGO_TOKEN_HMAC_SECRET_KEY", "change_me_centrifugo_token_secret"),
	}
}
