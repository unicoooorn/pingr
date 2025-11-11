package model

import (
    "fmt"
    "sort"
    "strings"
)

// это так выглядит подобие енамов в го :)
// https://dizzy.zone/2024/01/26/Enums-in-Go/
type PingStatus string

const (
	PingStatusOk    PingStatus = "ok"
	PingStatusNotOk PingStatus = "not_ok"
)

type Metric struct {
    Name   string
    Value  float64
    Labels map[string]string
}

type CheckResult struct {
    Status  PingStatus
    Metrics []Metric
    Details string // опционально, подробности возникновения статуса ("timeout 2.1s", "response: 502", и пр.)
}

func (m Metric) DumpToPrometheus() string {
    var sb strings.Builder
    sb.WriteString(m.Name)

    if len(m.Labels) > 0 {
        sb.WriteString("{")
        keys := make([]string, 0, len(m.Labels))
        for k := range m.Labels {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for i, k := range keys {
            if i > 0 {
                sb.WriteString(",")
            }
            fmt.Fprintf(&sb, `%s="%s"`, k, m.Labels[k])
        }
        sb.WriteString("}")
    }

    fmt.Fprintf(&sb, " %g\n", m.Value)
    return sb.String()
}