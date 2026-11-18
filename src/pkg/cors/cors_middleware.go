package cors_middleware

import (
	"net/http"
	"regexp"
	"strconv"
)

type corsMiddleware struct {
	regexArr []*regexp.Regexp
	config   *Config
}
type Middleware interface {
	Handler(handler http.Handler) http.Handler
}

// New erzeugt die CORS-Middleware aus der übergebenen Konfiguration.
//
// Die Origin-Regexes werden einmalig beim Start kompiliert, nicht pro
// Request. MustCompile ist hier Absicht: ein ungültiger Regex in der
// Konfiguration ist ein Deployment-Fehler und soll den Service sofort beim
// Start scheitern lassen, statt zur Laufzeit stillschweigend alle Origins
// abzulehnen.
//
// Hinweis zur Konfiguration: Die Regexes sollten mit ^ und $ verankert
// werden. Ein unverankertes Muster matcht als Teilstring – "huk.de" würde
// sonst auch "https://huk.de.angreifer.example" erlauben.
func New(corsConfig *Config) Middleware {
	regexArr := make([]*regexp.Regexp, len(corsConfig.CorsAllowOriginRegex))
	for i, regexStr := range corsConfig.CorsAllowOriginRegex {
		regexArr[i] = regexp.MustCompile(regexStr)
	}
	return &corsMiddleware{
		config:   corsConfig,
		regexArr: regexArr,
	}
}

func RegexSliceContains(slice []*regexp.Regexp, needle string) bool {
	for _, a := range slice {
		if a.MatchString(needle) {
			return true
		}
	}
	return false
}

func (c *corsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Es wird immer der konkrete Origin zurückgespiegelt, nie "*":
		// Ein Wildcard-Origin ist in Kombination mit
		// Access-Control-Allow-Credentials laut Fetch-Standard unzulässig
		// und würde vom Browser verworfen.
		if origin != "" && RegexSliceContains(c.regexArr, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Die erlaubten Origins werden per Regex geprüft, der Response
			// variiert also je Origin. Ohne "Vary: Origin" könnte ein Cache
			// die Antwort für einen fremden Origin ausliefern.
			w.Header().Add("Vary", "Origin")

			if c.config.CorsAllowMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", c.config.CorsAllowMethods)
			}
			if c.config.CorsAllowHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", c.config.CorsAllowHeaders)
			}
			if c.config.CorsAllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if c.config.CorsMaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.config.CorsMaxAge))
			}
		}

		// Preflight-Requests werden hier abschließend beantwortet und nicht
		// an die fachlichen Handler durchgereicht.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if next == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
