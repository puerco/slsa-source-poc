// SPDX-FileCopyrightText: Copyright 2025 The SLSA Authors
// SPDX-License-Identifier: Apache-2.0

package v1

import "encoding/json"

// Adds the control to the list. Ignores nil controls.
// Does not check for duplicate controls.
func (pred *Predicate) AddControl(control *Control) {
	if control == nil {
		return
	}
	pred.Controls = append(pred.Controls, control)
}

// Gets the control with the corresponding name, returns nil if not found.
func (pred *Predicate) GetControl(name string) *Control {
	for _, control := range pred.Controls {
		if control.Name == name {
			return control
		}
	}
	return nil
}

func (predicate *Predicate) MarshalJSON() ([]byte, error) {
	type Alias Predicate
	var createdOn string
	if predicate.CreatedOn != nil {
		createdOn = predicate.CreatedOn.AsTime().Format("2006-01-02T15:04:05.000Z")
	}

	return json.Marshal(
		&struct {
			Since string `json:"created_on,omitempty"`
			*Alias
		}{
			Since: createdOn,
			Alias: (*Alias)(predicate),
		},
	)
}
