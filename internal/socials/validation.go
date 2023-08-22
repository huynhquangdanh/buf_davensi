package socials

import (
	"errors"

	pbSocials "davensi.com/core/gen/socials"
)

// for Update gRPC
func validateQueryUpdate(msg *pbSocials.UpdateRequest) error {
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errors.New("id must be specified")
	}
	return nil
}
