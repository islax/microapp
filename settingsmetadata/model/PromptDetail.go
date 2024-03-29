package model

// PromptDetail contains the metadata regarding prompt for data change
type PromptDetail struct {
	PromptType           string        `json:"promptType"`
	PromptTitle          string        `json:"promptTitle"`
	PromptMessage        string        `json:"promptMessage"`
	ShowPromptWhenValues []interface{} `json:"showPromptWhenValues"` // If blank, shows prompt on every value change
}
