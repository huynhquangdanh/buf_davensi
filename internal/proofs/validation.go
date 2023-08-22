package proofs

import (
	"errors"

	pbProofs "davensi.com/core/gen/proofs"
)

func validateQuery(proof *pbProofs.Proof) error {
	if proof.GetId() == "" {
		return errors.New("id must be specified")
	}

	return nil
}
