package segment

import "time"

type Slug string

type Segment struct {
	Slug           Slug      `json:"slug" validate:"required"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`
}
