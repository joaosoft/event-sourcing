-- migrate up

-- :: AGGREGATE
CREATE SCHEMA IF NOT EXISTS eventsourcing;

CREATE TABLE eventsourcing.aggregate (
	id                        TEXT NOT NULL,
	type                      TEXT NOT NULL,
	version                   BIGINT NOT NULL,
	created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT aggregate_pkey PRIMARY KEY(id, type)
);

CREATE TABLE eventsourcing.snapshot (
	aggregate_id              TEXT NOT NULL,
	aggregate_type            TEXT NOT NULL,
	aggregate_version         BIGINT NOT NULL,
	data                      JSONB NOT NULL DEFAULT '{}',
	created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT snapshop_pkey PRIMARY KEY(aggregate_id, aggregate_type, aggregate_version),
	CONSTRAINT snapshop_aggregate_fkey FOREIGN KEY(aggregate_id, aggregate_type) REFERENCES eventsourcing.aggregate(id, type) INITIALLY DEFERRED
);

CREATE TABLE eventsourcing.event (
	id                        TEXT NOT NULL UNIQUE,
	name                      TEXT NOT NULL,
	aggregate_id              TEXT NOT NULL,
	aggregate_type            TEXT NOT NULL,
	aggregate_version         BIGINT NOT NULL,
	data                      JSONB DEFAULT '{}',
	created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT event_pkey PRIMARY KEY(id),
	CONSTRAINT event_aggregate_fkey FOREIGN KEY(aggregate_id, aggregate_type) REFERENCES eventsourcing.aggregate(id, type) INITIALLY DEFERRED
);





-- :: WEBHOOK HANDLING

CREATE TABLE eventsourcing.webhook (
	id                        TEXT NOT NULL,
	aggregate_type            TEXT NOT NULL,
	name                      TEXT NOT NULL,
	callback                  TEXT NOT NULL,
	authentication            TEXT,
	token                     TEXT,
	"user"                    TEXT,
	password                  TEXT,
	created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT webhook_pkey PRIMARY KEY(id)
);

CREATE TABLE eventsourcing.dispatch (
	webhook_id                TEXT NOT NULL,
  event_id                  TEXT NOT NULL,
  event_name                TEXT NOT NULL,
	aggregate_id              TEXT NOT NULL,
	aggregate_type            TEXT NOT NULL,
	aggregate_version         BIGINT NOT NULL,
	data                      JSONB DEFAULT '{}',
	is_dispatched             BOOLEAN NOT NULL DEFAULT FALSE,
	created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT dispatch_pkey PRIMARY KEY (webhook_id, event_id),
	CONSTRAINT dispatch_event_fkey FOREIGN KEY(event_id) REFERENCES eventsourcing.event(id) INITIALLY deferred,
	CONSTRAINT dispatch_webhook_fkey FOREIGN KEY(webhook_id) REFERENCES eventsourcing.webhook(id) INITIALLY deferred,
	CONSTRAINT dispatch_aggregate_fkey FOREIGN KEY(aggregate_id, aggregate_type) REFERENCES eventsourcing.aggregate(id, type) INITIALLY deferred
);





-- migrate down

DROP TABLE eventsourcing.aggregate;
DROP TABLE eventsourcing.snapshot;
DROP TABLE eventsourcing.event;

DROP TABLE eventsourcing.webhook;
DROP TABLE eventsourcing.dispatch;
