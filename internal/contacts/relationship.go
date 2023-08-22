package contacts

import (
	"context"

	pbContacts "davensi.com/core/gen/contacts"
	"github.com/rs/zerolog/log"
)

func (s *ServiceServer) GetListInternal(ctx context.Context, req *pbContacts.GetListRequest,
) (contactList []*pbContacts.Contact, err error) {
	qb := s.Repo.QbGetList(req)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		contact, err := s.Repo.ScanRow(rows)
		if err != nil {
			return nil, err
		}

		contactList = append(contactList, contact)
	}

	return contactList, nil
}
