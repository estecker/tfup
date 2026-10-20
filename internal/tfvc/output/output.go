package output

import "sort"

type format string
type Format format

const (
    FormatJSON Format = "json"
)

var (
    formats = map[string]Format{
        string(FormatJSON): FormatJSON,
    }
    FormatNames = make([]string, 0, len(formats))
)

func init() {
    for k := range formats {
        FormatNames = append(FormatNames, k)
    }
    sort.Strings(FormatNames)
}

func ParseFormatName(s string) (Format, bool) {
    f, ok := formats[s]
    return f, ok
}
