------------------------------------------------------------
-- DROP TABLES (dev only — safe to remove for production)
------------------------------------------------------------
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS appdata CASCADE;
DROP TABLE IF EXISTS applications CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS statuses CASCADE;
DROP TABLE IF EXISTS sources CASCADE;
DROP TABLE IF EXISTS eventtypes CASCADE;

------------------------------------------------------------
-- LOOKUP TABLES
------------------------------------------------------------

CREATE TABLE statuses (
    id         SERIAL PRIMARY KEY,
    label      VARCHAR(16) NOT NULL,
    listorder  INT NOT NULL DEFAULT 0,
    active     BOOLEAN NOT NULL
);

CREATE TABLE sources (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    url        TEXT,
    listorder  INT NOT NULL DEFAULT 0,
    active     BOOLEAN NOT NULL
);

CREATE TABLE eventtypes (
    id         SERIAL PRIMARY KEY,
    label      TEXT NOT NULL,
    listorder  INT NOT NULL DEFAULT 0,
    active     BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS public.sessions
(
    sessionid text COLLATE pg_catalog."default" NOT NULL,
    userid integer NOT NULL,
    created timestamp with time zone NOT NULL DEFAULT now(),
    expires timestamp with time zone NOT NULL,
    active boolean NOT NULL DEFAULT true,
    CONSTRAINT sessions_pkey PRIMARY KEY (sessionid),
    CONSTRAINT sessions_userid_fkey FOREIGN KEY (userid)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
)


------------------------------------------------------------
-- INSERT "NONE" ROWS (id = 0)
------------------------------------------------------------

INSERT INTO statuses (id, label, listorder, active)
VALUES (0, 'None', 0, TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO sources (id, name, listorder, active)
VALUES (0, 'None', 0, TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO eventtypes (id, label, listorder, active)
VALUES (0, 'None', 0, TRUE)
ON CONFLICT DO NOTHING;

------------------------------------------------------------
-- USERS
------------------------------------------------------------

CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    name          TEXT,
    username      TEXT NOT NULL,
    passwordhash  TEXT NOT NULL,
    created       TIMESTAMPTZ DEFAULT now(),
    updated       TIMESTAMPTZ DEFAULT now(),
    lastlogin     TIMESTAMPTZ
);

------------------------------------------------------------
-- ORGANIZATIONS
------------------------------------------------------------

CREATE TABLE organizations (
    id        SERIAL PRIMARY KEY,
    userid    INT NOT NULL,
    name      TEXT NOT NULL,
    url       TEXT,
    created   TIMESTAMPTZ DEFAULT now(),
    updated   TIMESTAMPTZ DEFAULT now()
);

------------------------------------------------------------
-- APPLICATIONS
------------------------------------------------------------

CREATE TABLE applications (
    id              SERIAL PRIMARY KEY,
    userid          INT NOT NULL,
    organizationid  INT,
    position        TEXT,
    dateapplied     TIMESTAMPTZ,
    lastresponse    TIMESTAMPTZ,
    statusid        INT NOT NULL DEFAULT 0,
    jobposting      TEXT,
    notes           TEXT,
    url             TEXT,
    siteuser        TEXT,
    sitepass        TEXT,
    sourceid        INT,
    created         TIMESTAMPTZ DEFAULT now(),
    updated         TIMESTAMPTZ DEFAULT now()
);

------------------------------------------------------------
-- CONTACTS
------------------------------------------------------------

CREATE TABLE contacts (
    id              SERIAL PRIMARY KEY,
    userid          INT NOT NULL,
    organizationid  INT,
    name            TEXT,
    email           TEXT,
    notes           TEXT,
    created         TIMESTAMPTZ DEFAULT now(),
    updated         TIMESTAMPTZ DEFAULT now()
);

------------------------------------------------------------
-- EVENTS
------------------------------------------------------------

CREATE TABLE events (
    id              SERIAL PRIMARY KEY,
    userid          INT NOT NULL,
    applicationid   INT,
    organizationid  INT,
    eventtypeid     INT NOT NULL DEFAULT 0,
    name            TEXT,
    date            TIMESTAMPTZ,
    notes           TEXT,
    created         TIMESTAMPTZ DEFAULT now(),
    updated         TIMESTAMPTZ DEFAULT now()
);

------------------------------------------------------------
-- APPDATA
------------------------------------------------------------

CREATE TABLE appdata (
    id        SERIAL PRIMARY KEY,
    userid    INT NOT NULL,
    subject   TEXT,
    content   TEXT,
    created   TIMESTAMPTZ DEFAULT now(),
    updated   TIMESTAMPTZ DEFAULT now()
);

------------------------------------------------------------
-- FOREIGN KEYS
------------------------------------------------------------

-- USERS OWN EVERYTHING
ALTER TABLE organizations
    ADD CONSTRAINT fk_org_user
    FOREIGN KEY (userid) REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE applications
    ADD CONSTRAINT fk_app_user
    FOREIGN KEY (userid) REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE contacts
    ADD CONSTRAINT fk_contact_user
    FOREIGN KEY (userid) REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE events
    ADD CONSTRAINT fk_event_user
    FOREIGN KEY (userid) REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE appdata
    ADD CONSTRAINT fk_appdata_user
    FOREIGN KEY (userid) REFERENCES users(id)
    ON DELETE CASCADE;

-- SOFT REFERENCES
ALTER TABLE applications
    ADD CONSTRAINT fk_app_org
    FOREIGN KEY (organizationid) REFERENCES organizations(id)
    ON DELETE SET NULL;

ALTER TABLE events
    ADD CONSTRAINT fk_event_app
    FOREIGN KEY (applicationid) REFERENCES applications(id)
    ON DELETE SET NULL;

ALTER TABLE events
    ADD CONSTRAINT fk_event_org
    FOREIGN KEY (organizationid) REFERENCES organizations(id)
    ON DELETE SET NULL;

-- LOOKUP TABLES (RESTRICT)
ALTER TABLE applications
    ADD CONSTRAINT fk_app_status
    FOREIGN KEY (statusid) REFERENCES statuses(id)
    ON DELETE RESTRICT;

ALTER TABLE applications
    ADD CONSTRAINT fk_app_source
    FOREIGN KEY (sourceid) REFERENCES sources(id)
    ON DELETE RESTRICT;

ALTER TABLE events
    ADD CONSTRAINT fk_event_type
    FOREIGN KEY (eventtypeid) REFERENCES eventtypes(id)
    ON DELETE RESTRICT;

------------------------------------------------------------
-- TIMESTAMP TRIGGER FUNCTION
------------------------------------------------------------

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

------------------------------------------------------------
-- TRIGGERS
------------------------------------------------------------

CREATE TRIGGER trg_users_timestamp
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_org_timestamp
BEFORE UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_app_timestamp
BEFORE UPDATE ON applications
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_contact_timestamp
BEFORE UPDATE ON contacts
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_event_timestamp
BEFORE UPDATE ON events
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_appdata_timestamp
BEFORE UPDATE ON appdata
FOR EACH ROW EXECUTE FUNCTION update_timestamp();