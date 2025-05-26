// SPDX-FileCopyrightText: Copyright 2025 The SLSA Authors
// SPDX-License-Identifier: Apache-2.0

package v1

import "encoding/json"

type Controls []*Control

// Adds the control to the list. Ignores nil controls.
// Does not check for duplicate controls.
func (controls *Controls) AddControl(control *Control) {
	if control == nil {
		return
	}
	*controls = append(*controls, control)
}

// Gets the control with the corresponding name, returns nil if not found.
func (controls *Controls) GetControl(name string) *Control {
	for _, control := range *controls {
		if control.Name == name {
			return control
		}
	}
	return nil
}

func (control *Control) MarshalJSON() ([]byte, error) {
	type Alias Control
	var since string
	if control.Since != nil {
		since = control.Since.AsTime().Format("2006-01-02T15:04:05.000Z")
	}

	return json.Marshal(
		&struct {
			Since string `json:"since,omitempty"`
			*Alias
		}{
			Since: since,
			Alias: (*Alias)(control),
		},
	)
}
