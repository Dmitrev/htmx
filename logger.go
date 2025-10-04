package main

import (
	"fmt"
	"strings"
)


type Color string

const (
	Reset  Color = "\033[0m"
	Red    Color = "\033[31m"
	Green  Color = "\033[32m"
	Yellow Color = "\033[33m"
	Blue   Color = "\033[34m"
)

const tableWidth int = 100

type Logger struct {
}

func NewLogger() Logger {
    return Logger{}
}

func (l *Logger) Sep() {
    fmt.Println(strings.Repeat("-", tableWidth))
}
func (l *Logger) PrintHeading() {
    l.Sep()
    fmt.Println("| Method | Uri                                                                                     |")
    l.Sep()
}

func (l *Logger) Col(content string, length int, color Color) string {

    paddingLeft := 0
    paddingRight := 0

    if len(content) < length {
	paddingNeeded := length - len(content)
	// standard padding
	paddingLeft++
	paddingNeeded--
	paddingRight += paddingNeeded
	paddingNeeded = 0
    }

    builder := strings.Builder{}
    builder.WriteString(string(color))
    builder.WriteString(strings.Repeat(" ", paddingLeft))
    builder.WriteString(content)
    builder.WriteString(strings.Repeat(" ", paddingRight))
    builder.WriteString(string(Reset))

    return builder.String()
}

func (l *Logger) PrintRequest(method, uri string) {

    color := Reset
    switch (method) {
    case "GET": color = Green
    case "POST": color = Blue
    case "PUT": color = Yellow
    case "PATCH": color = Red
    }

    methodColumnWidth := 8
    colCount := 2
    builder := strings.Builder{}
    builder.WriteString("|")
    builder.WriteString(l.Col(method, methodColumnWidth, color))
    builder.WriteString("|")
    builder.WriteString(l.Col(uri, tableWidth - methodColumnWidth - (colCount + 1), Reset))
    builder.WriteString("|")
    fmt.Println(builder.String())

    l.Sep()

    builder.Reset()
    builder.WriteString("|")
    builder.WriteString("|")
    fmt.Println(builder.String())
}

