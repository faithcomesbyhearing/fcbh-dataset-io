package uroman

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"os"
	"path/filepath"
)

// uroman.go requires a pip install uroman, but it uses the uroman.pl that is included
// https://github.com/isi-nlp/uroman/tree/master

func EnsureUroman(conn db.DBAdapter, lang string) *log.Status {
	hasUroman, status := CheckUroman(conn)
	if status != nil {
		return status
	}
	if !hasUroman {
		status = UpdateUroman(conn, lang)
	}
	return status
}

func CheckUroman(conn db.DBAdapter) (bool, *log.Status) {
	var result bool
	textLen, status := conn.SelectScalarInt("SELECT sum(length(script_text)) FROM scripts")
	if status != nil {
		return result, status
	}
	uromanLen, status := conn.SelectScalarInt("SELECT sum(length(uroman)) FROM scripts")
	if status != nil {
		return result, status
	}
	return float64(uromanLen)*1.2 >= float64(textLen), status
}

func UpdateUroman(conn db.DBAdapter, lang string) *log.Status {
	scripts, status := conn.SelectScripts()
	if status != nil {
		return status
	}
	scripts, status = SetUroman(conn.Ctx, scripts, lang)
	if status != nil {
		return status
	}
	_, status = conn.UpdateUromanText(scripts)
	return status
}

func SetUroman(ctx context.Context, linesIn []db.Script, lang string) (lines []db.Script, status *log.Status) {
	lines = linesIn
	uroman, status := stdio_exec.NewStdioExec(ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), ScriptPath(), "-l", lang)
	if status != nil {
		return lines, status
	}
	defer func() {
		status = uroman.Close()
	}()
	for i := range lines {
		lines[i].URoman, status = uroman.Process(lines[i].ScriptText)
		if status != nil {
			return lines, status
		}
	}
	return lines, status
}

func ScriptPath() string {
	return filepath.Join(os.Getenv("GOPROJ"), "utility", "uroman", "uroman_stdio.py")
}
