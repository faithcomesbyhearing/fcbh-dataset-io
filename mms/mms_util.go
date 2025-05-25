package mms

import (
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/lang_tree/search"
	"strings"
)

// Check that language is supported by mms_asr, and return alternate if it is not
func CheckLanguage(ctx context.Context, lang string, sttLang string, aiTool string) (string, *log.Status) {
	var result string
	if sttLang != `` {
		result = sttLang
	} else {
		var tree = search.NewLanguageTree(ctx)
		err := tree.Load()
		if err != nil {
			return result, log.Error(ctx, 500, err, `Error loading language`)
		}
		langs, distance, err2 := tree.Search(strings.ToLower(lang), aiTool)
		if err2 != nil {
			return result, log.Error(ctx, 500, err2, `Error Searching for language`)
		}
		if len(langs) > 0 {
			result = langs[0]
			log.Info(ctx, `Using language`, result, "distance:", distance)
		} else {
			return result, log.ErrorNoErr(ctx, 400, `No compatible language code was found for`, lang)
		}
	}
	return result, nil
}
