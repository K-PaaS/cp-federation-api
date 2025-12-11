package localize

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log/slog"
	"os"
)

var Localizer *i18n.Localizer
var Bundle *i18n.Bundle

func init() {
	LocalizeConfig()
}

func LocalizeConfig() {
	cwd, _ := os.Getwd()
	slog.Info("cwd", "localize/localize.go", cwd)
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	Bundle.LoadMessageFile("app/localize/en.json")
	Bundle.LoadMessageFile("app/localize/ko.json")

}

func GetLocalizeMessage(c *gin.Context, MessageID string) string {
	lang := c.GetString("lang")
	Localizer = i18n.NewLocalizer(Bundle, lang)

	localizeConfig := i18n.LocalizeConfig{
		MessageID: MessageID,
	}
	localizationUsingJson, err := Localizer.Localize(&localizeConfig)
	if err != nil {
		return MessageID
	}
	return localizationUsingJson
}
