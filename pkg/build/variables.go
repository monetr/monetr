package build

import "embed"

var (
	Revision  string
	BuildTime string
	BuildHost string
	BuildType string
	Release   string
)

//go:embed *.md
var noticeFS embed.FS

func GetNotice() string {
	data, err := noticeFS.ReadFile("NOTICE.md")
	if err != nil {
		return ""
	}
	return string(data)
}
