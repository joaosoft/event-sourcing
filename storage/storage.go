package storage

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"event-sourcing/models"
	logger "github.com/joaosoft/logger"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (storage *Storage) GetAggregate(id, typ string) (*models.Aggregate, error) {

	// aggregate
	var aggregate = models.Aggregate{Id: id, Type: typ}
	row := storage.db.QueryRow(`
		SELECT version, created_at, updated_at
		FROM eventsourcing.aggregate
		WHERE id = $1 AND type = $2
	`, aggregate.Id, aggregate.Type)

	if err := row.Scan(&aggregate.Version, &aggregate.CreatedAt, &aggregate.UpdatedAt); err != nil && err != sql.ErrNoRows {
		logger.WithField("error", err.Error()).Error("error getting aggregate from database")
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	// snapshot
	aggregate.Data = make([]byte, 0)
	row = storage.db.QueryRow(`
		SELECT data, created_at
		FROM eventsourcing.snapshot
		WHERE aggregate_id = $1 AND aggregate_type = $2 AND aggregate_version = $3
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	if err := row.Scan(&aggregate.Data, &aggregate.CreatedAt); err != nil {
		logger.WithField("error", err.Error()).Error("error getting snapshot from database")
		return nil, err
	}

	// events
	rows, err := storage.db.Query(`
		SELECT id, data, created_at
		FROM eventsourcing.event
		WHERE aggregate_id = $1 AND aggregate_type = $2 AND aggregate_version = $3
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	if err != nil {
		logger.WithField("error", err.Error()).Error("error getting events from database")
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var event models.Event
		var id string
		if err := rows.Scan(&id, &event.Data, &event.CreatedAt); err != nil {
			logger.WithField("error", err.Error()).Error("error getting event from database")
			return nil, err
		}
		event.Id, err = ulid.Parse(id)
		if err != nil {
			logger.WithField("error", err.Error()).Error("error converting ulid from database")
			return nil, err
		}
		aggregate.Events = append(aggregate.Events, &event)
	}

	return &aggregate, nil
}

func (storage *Storage) StoreAggregate(aggregate *models.Aggregate) (err error) {

	if len(aggregate.Events) == 0 {
		var err error
		logger.Error("there is no event on the aggregate").ToError(&err)
		return err
	}

	var aggregateData = &bytes.Buffer{}
	if err := json.NewEncoder(aggregateData).Encode(aggregate.Data); err != nil {
		logger.WithField("error", err.Error()).Error("error encoding aggregate data")
		return err
	}

	tx, err := storage.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				logger.WithField("error", err.Error()).Error("doing rollback of transaction on database")
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}
	}()

	row := tx.QueryRow(`
		SELECT version
		FROM eventsourcing.aggregate
		WHERE id = $1 AND type = $2
		FOR UPDATE NOWAIT
	`, aggregate.Id, aggregate.Type)

	if err := row.Scan(&aggregate.Version); err != nil && err != sql.ErrNoRows {
		logger.WithField("error", err.Error()).Error("error getting aggregate from database")
		return err
	}

	// aggregate
	if aggregate.Version == 0 {
		aggregate.Version += 1
		_, err = tx.Exec(`
		INSERT INTO eventsourcing.aggregate (id,  type, version)
		VALUES($1, $2, $3)
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	} else {
		aggregate.Version += 1
		_, err = tx.Exec(`
		UPDATE eventsourcing.aggregate
		SET version = $1, updated_at = now()
		WHERE id = $2 AND type = $3
	`, aggregate.Version, aggregate.Id, aggregate.Type)
	}
	if err != nil {
		logger.WithField("error", err.Error()).Error("error inserting/updating aggregate on database")
		return err
	}

	// events
	stmt, err := tx.Prepare(`
		INSERT INTO eventsourcing.event (id, name, aggregate_id, aggregate_type, aggregate_version, data)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		logger.WithField("error", err.Error()).Error("error inserting event from database")
		return err
	}

	defer stmt.Close()
	for _, newEvent := range aggregate.Events {
		var eventData = &bytes.Buffer{}
		if err := json.NewEncoder(eventData).Encode(newEvent.GetData()); err != nil {
			logger.WithField("error", err.Error()).Error("error encoding event data")
			return err
		}
		if _, err = stmt.Exec(newEvent.GetId().String(), newEvent.GetName(), aggregate.Id, aggregate.Type, aggregate.Version, eventData.Bytes()); err != nil {
			logger.WithField("error", err.Error()).Error("error inserting event on database")
			return err
		}
	}

	// snapshot
	if _, err = tx.Exec(`
		INSERT INTO eventsourcing.snapshot (aggregate_id, aggregate_type, aggregate_version, data)
		VALUES($1, $2, $3, $4)
	`, aggregate.Id, aggregate.Type, aggregate.Version, aggregateData.Bytes()); err != nil {
		logger.WithField("error", err.Error()).Error("error inserting snapshot on database")
		return err
	}

	// dispatch
	rows, err := storage.db.Query(`
		SELECT id
		FROM eventsourcing.webhook
		WHERE aggregate_type = $1
	`, aggregate.Type)
	if err != nil {
		logger.WithField("error", err.Error()).Error("error getting webhooks from database")
		return err
	}

	defer rows.Close()
	var webhookID string
	for rows.Next() {
		if err := rows.Scan(&webhookID); err != nil {
			logger.WithField("error", err.Error()).Error("error getting webhook from database")
			return err
		}

		for _, event := range aggregate.Events {
			var eventData = &bytes.Buffer{}
			if err := json.NewEncoder(eventData).Encode(event.GetData()); err != nil {
				logger.WithField("error", err.Error()).Error("error encoding event data")
				return err
			}

			if _, err = tx.Exec(`
				INSERT INTO eventsourcing.dispatch (event_id, event_name, webhook_id, aggregate_id, aggregate_type, aggregate_version, data)
				VALUES($1, $2, $3, $4, $5, $6, $7)
			`, event.GetId().String(), event.GetName(), webhookID, aggregate.Id, aggregate.Type, aggregate.Version, eventData.Bytes()); err != nil {
				logger.WithField("error", err.Error()).Error("error inserting dispatch on database")
				return err
			}
		}
	}

	return nil
}
