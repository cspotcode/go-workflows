package sqlite

import (
	"context"
	"database/sql"

	"github.com/cschleiden/go-dt/pkg/history"
)

func scheduleActivity(ctx context.Context, tx *sql.Tx, instanceID, executionID string, event history.Event) error {
	attributes, err := history.SerializeAttributes(event.Attributes)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO activities
			(id, instance_id, execution_id, event_type, event_id, attributes, visible_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		instanceID,
		executionID,
		event.Type,
		event.EventID,
		attributes,
		event.VisibleAt,
	)

	return err
}