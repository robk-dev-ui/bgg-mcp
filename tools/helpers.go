package tools

import (
	"strings"

	"github.com/kkjdaniel/gogeek/thing"
)

type EssentialGameInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Year        int     `json:"year"`
	Complexity  float64 `json:"complexity"`
	Players     string  `json:"players"`
	BGGRating   float64 `json:"bgg_rating"`
	PlayTime    string  `json:"play_time"`
	Designer    string  `json:"designer"`
}

func extractEssentialInfo(item thing.Item) EssentialGameInfo {
	info := EssentialGameInfo{
		Name: item.Name.Value,
		Year: item.YearPublished.Value,
	}

	// Description
	if len(item.Description) > 0 {
		desc := item.Description[0].Value
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		info.Description = desc
	}

	// Complexity
	if item.Statistics != nil && item.Statistics.AverageWeight.Value > 0 {
		info.Complexity = item.Statistics.AverageWeight.Value
	}

	// Players
	if item.MinPlayers.Value > 0 && item.MaxPlayers.Value > 0 {
		if item.MinPlayers.Value == item.MaxPlayers.Value {
			info.Players = string(rune(item.MinPlayers.Value + '0'))
		} else {
			info.Players = string(rune(item.MinPlayers.Value+'0')) + "-" + string(rune(item.MaxPlayers.Value+'0'))
		}
	}

	// BGG Rating
	if item.Statistics != nil && item.Statistics.Average.Value > 0 {
		info.BGGRating = item.Statistics.Average.Value
	}

	// Play Time
	if item.MinPlayTime.Value > 0 && item.MaxPlayTime.Value > 0 {
		if item.MinPlayTime.Value == item.MaxPlayTime.Value {
			info.PlayTime = string(rune(item.MinPlayTime.Value/10+'0')) + "0 min"
		} else {
			info.PlayTime = string(rune(item.MinPlayTime.Value/10+'0')) + "0-" + string(rune(item.MaxPlayTime.Value/10+'0')) + "0 min"
		}
	}

	// Designer
	var designers []string
	for _, link := range item.Link {
		if link.Type == "boardgamedesigner" {
			designers = append(designers, link.Value)
		}
	}
	if len(designers) > 0 {
		info.Designer = strings.Join(designers, ", ")
	}

	return info
}

func extractEssentialInfoList(items []thing.Item) []EssentialGameInfo {
	result := make([]EssentialGameInfo, len(items))
	for i, item := range items {
		result[i] = extractEssentialInfo(item)
	}
	return result
}