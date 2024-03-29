package controller

import (
	"k8s.io/utils/strings/slices"
	"strings"
)

func int32Ptr(i int32) *int32 { return &i }

func map2String(kv map[string]string, linePrefixSpaces int, skipList []string) string {
	var sb strings.Builder
	for key, value := range kv {
		if skipList != nil {
			if slices.Contains(skipList, key) {
				continue
			}
		}
		writeSpaces(&sb, linePrefixSpaces)
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

func list2String(l []string) string {
	var sb strings.Builder
	for index, value := range l {
		sb.WriteString(value)
		if index < (len(l) - 1) {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func writeSpaces(sb *strings.Builder, count int) {
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		sb.WriteString(" ")
	}
}
