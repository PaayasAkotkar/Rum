package rumpaint

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

type pallete struct {
	A1 string
}
type rumColors struct {
	Green pallete
	Blue  pallete
	Black pallete
	White pallete
}

func colors() rumColors {
	return rumColors{
		Green: pallete{
			A1: "#D0F5AB",
		},
		Blue: pallete{
			A1: "#ABEAF5",
		},
		Black: pallete{
			A1: "#111111",
		},
		White: pallete{
			A1: "#ffffff",
		},
	}
}
func Header(header string) string {
	p1 := lipgloss.NewStyle().Foreground(lipgloss.Color(colors().Green.A1)).Bold(true).Align(lipgloss.Center).Render(header)
	p := lipgloss.NewStyle().
		Padding(0, 2).
		Render(p1)
	return p
}
func Title(title string) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors().Black.A1)).
		Background(lipgloss.Color(colors().Blue.A1)).
		Padding(1, 10).
		MarginTop(1).
		MarginBottom(1).
		Align(lipgloss.Center).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(colors().Green.A1))

	titleStr := strings.ToUpper(title)
	renderedHeader := headerStyle.Render(titleStr)

	return renderedHeader
}

func Table(title string, headers []string, data [][]string) *table.Table {
	CapitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}
	text := lipgloss.NewStyle().
		Background(lipgloss.Color(colors().Green.A1)).
		Foreground(lipgloss.Color(colors().Black.A1)).
		Padding(0, 2). // Top/Bottom 0, Left/Right 2
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors().Blue.A1)). // Match border to background for a "pill" look
		Align(lipgloss.Center).
		Width(20)
	lipgloss.Println(text.Render(title))

	headerStyle := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors().Green.A1)).Bold(true).Align(lipgloss.Center)
	pencil := table.New().Border(lipgloss.NormalBorder()).BorderStyle(borderStyle).Headers(CapitalizeHeaders(headers)...).Width(150).Rows(data...).StyleFunc(func(row, col int) lipgloss.Style {

		if row == table.HeaderRow {
			return headerStyle
		}
		return lipgloss.NewStyle().Align(lipgloss.Center)
	})
	return pencil
}

func Card(title, desc string) string {
	pencil := lipgloss.NewStyle().
		Width(20).
		Height(5).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colors().Green.A1)).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors().Green.A1))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors().White.A1))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		descStyle.Render(desc),
	)

	p := pencil.Render(content)
	return p
}
