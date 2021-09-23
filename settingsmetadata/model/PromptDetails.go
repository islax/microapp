package model

// PromptDetails contains the metadata regarding prompt for data change
type PromptDetails struct {
	PromptTitle          string        `json:"promptTitle"`
	PromptMessage        string        `json:"promptMessage"`
	ShowPromptWhenValues []interface{} `json:"showPromptWhenValues"` // If blank, shows prompt on every value change
}
