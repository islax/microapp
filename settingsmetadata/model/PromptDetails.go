package model

// PromptDetails contains the metadata regarding prompt for data change
type PromptDetails struct {
	PromptTitle                 string        `json:"promptTitle"`
	PromptMessage               string        `json:"promptMessage"`
	ShowPromptOnWhichDataChange []interface{} `json:"showPromptOnWhichDataChange"`
}
