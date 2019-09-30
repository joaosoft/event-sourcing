package models

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/joaosoft/logger"

	_ "github.com/lib/pq"
	"github.com/oklog/ulid"
)

type Storage struct {
	db     *sql.DB
	logger logger.ILogger
}

func NewStorage(db *sql.DB, logger logger.ILogger) *Storage {
	return &Storage{db: db, logger: logger}
}

func (s *Storage) GetAggregate(id, typ string, obj interface{}) (*Aggregate, error) {

	// aggregate
	var aggregate = Aggregate{Id: id, Type: typ}
	row := s.db.QueryRow(`
		SELECT version, created_at, updated_at
		FROM eventsourcing.aggregate
		WHERE id = $1 AND type = $2
	`, aggregate.Id, aggregate.Type)

	if err := row.Scan(&aggregate.Version, &aggregate.CreatedAt, &aggregate.UpdatedAt); err != nil && err != sql.ErrNoRows {
		s.logger.WithField("error", err.Error()).Error("error getting aggregate from database")
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	// snapshot
	aggregateBytes := make([]byte, 0)
	row = s.db.QueryRow(`
		SELECT data, created_at
		FROM eventsourcing.snapshot
		WHERE aggregate_id = $1 AND aggregate_type = $2 AND aggregate_version = $3
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	if err := row.Scan(&aggregateBytes, &aggregate.CreatedAt); err != nil {
		s.logger.WithField("error", err.Error()).Error("error getting snapshot from database")
		return nil, err
	}

	// events
	rows, err := s.db.Query(`
		SELECT id, data, created_at
		FROM eventsourcing.event
		WHERE aggregate_id = $1 AND aggregate_type = $2 AND aggregate_version = $3
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("error getting events from database")
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var event Event
		var id string
		if err := rows.Scan(&id, &event.Data, &event.CreatedAt); err != nil {
			s.logger.WithField("error", err.Error()).Error("error getting event from database")
			return nil, err
		}
		event.Id, err = ulid.Parse(id)
		if err != nil {
			s.logger.WithField("error", err.Error()).Error("error converting ulid from database")
			return nil, err
		}
		aggregate.Events = append(aggregate.Events, &event)
	}

	err = json.Unmarshal(aggregateBytes, obj)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("error unmarshal aggregate data")
		return nil, err
	}
	aggregate.Data = obj

	return &aggregate, nil
}

func (s *Storage) StoreAggregate(aggregate *Aggregate) (err error) {

	if len(aggregate.Events) == 0 {
		return s.logger.Error("there is no event on the aggregate").ToError()
	}

	var aggregateData = &bytes.Buffer{}
	if err := json.NewEncoder(aggregateData).Encode(aggregate.Data); err != nil {
		s.logger.WithField("error", err.Error()).Error("error encoding aggregate data")
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if tx != nil {
			if err != nil {
				s.logger.WithField("error", err.Error()).Error("doing rollback of transaction on database")
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
		s.logger.WithField("error", err.Error()).Error("error getting aggregate from database")
		return err
	}

	// aggregate
	aggregate.Version += 1

	if aggregate.Version == 0 {
		_, err = tx.Exec(`
		INSERT INTO eventsourcing.aggregate (id,  type, version)
		VALUES($1, $2, $3)
	`, aggregate.Id, aggregate.Type, aggregate.Version)
	} else {
		_, err = tx.Exec(`
		UPDATE eventsourcing.aggregate
		SET version = $1, updated_at = now()
		WHERE id = $2 AND type = $3
	`, aggregate.Version, aggregate.Id, aggregate.Type)
	}
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("error inserting/updating aggregate on database")
		return err
	}

	// events
	stmt, err := tx.Prepare(`
		INSERT INTO eventsourcing.event (id, name, aggregate_id, aggregate_type, aggregate_version, data)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("error inserting event from database")
		return err
	}

	defer stmt.Close()
	for _, newEvent := range aggregate.Events {
		var eventData = &bytes.Buffer{}
		if err := json.NewEncoder(eventData).Encode(newEvent.GetData()); err != nil {
			s.logger.WithField("error", err.Error()).Error("error encoding event data")
			return err
		}
		if _, err = stmt.Exec(newEvent.GetId().String(), newEvent.GetName(), aggregate.Id, aggregate.Type, aggregate.Version, eventData.Bytes()); err != nil {
			s.logger.WithField("error", err.Error()).Error("error inserting event on database")
			return err
		}
	}

	// snapshot
	if _, err = tx.Exec(`
		INSERT INTO eventsourcing.snapshot (aggregate_id, aggregate_type, aggregate_version, data)
		VALUES($1, $2, $3, $4)
	`, aggregate.Id, aggregate.Type, aggregate.Version, aggregateData.Bytes()); err != nil {
		s.logger.WithField("error", err.Error()).Error("error inserting snapshot on database")
		return err
	}

	// dispatch
	rows, err := s.db.Query(`
		SELECT id
		FROM eventsourcing.webhook
		WHERE aggregate_type = $1
	`, aggregate.Type)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("error getting webhooks from database")
		return err
	}

	defer rows.Close()
	var webhookID string
	for rows.Next() {
		if err := rows.Scan(&webhookID); err != nil {
			s.logger.WithField("error", err.Error()).Error("error getting webhook from database")
			return err
		}

		for _, event := range aggregate.Events {
			var eventData = &bytes.Buffer{}
			if err := json.NewEncoder(eventData).Encode(event.GetData()); err != nil {
				s.logger.WithField("error", err.Error()).Error("error encoding event data")
				return err
			}

			if _, err = tx.Exec(`
				INSERT INTO eventsourcing.dispatch (event_id, event_name, webhook_id, aggregate_id, aggregate_type, aggregate_version, data)
				VALUES($1, $2, $3, $4, $5, $6, $7)
			`, event.GetId().String(), event.GetName(), webhookID, aggregate.Id, aggregate.Type, aggregate.Version, eventData.Bytes()); err != nil {
				s.logger.WithField("error", err.Error()).Error("error inserting dispatch on database")
				return err
			}
		}
	}

	return nil
}
