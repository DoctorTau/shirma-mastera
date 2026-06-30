CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE monsters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dndsu_id int NOT NULL,
    slug text NOT NULL,
    edition text NOT NULL CHECK (edition IN ('2014', '2024')),
    name_ru text NOT NULL DEFAULT '',
    name_en text NOT NULL DEFAULT '',
    cr text NOT NULL DEFAULT '',
    cr_numeric numeric,
    type text NOT NULL DEFAULT '',
    size text NOT NULL DEFAULT '',
    alignment text NOT NULL DEFAULT '',
    statblock jsonb NOT NULL DEFAULT '{}'::jsonb,
    image_url text,
    source_book text,
    source_url text NOT NULL,
    raw_html text,
    content_hash text NOT NULL DEFAULT '',
    linked_monster_id uuid REFERENCES monsters(id),
    is_unique_npc boolean NOT NULL DEFAULT false,
    last_fetched_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (slug, edition)
);

CREATE INDEX idx_monsters_name_ru ON monsters USING gin (to_tsvector('russian', name_ru));
CREATE INDEX idx_monsters_name_en ON monsters USING gin (to_tsvector('simple', name_en));
CREATE INDEX idx_monsters_cr ON monsters (cr_numeric);
CREATE INDEX idx_monsters_type ON monsters (type);
CREATE INDEX idx_monsters_edition ON monsters (edition);
CREATE INDEX idx_monsters_dndsu_id ON monsters (dndsu_id);

CREATE TABLE catalog_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    section text NOT NULL DEFAULT 'monster',
    edition text NOT NULL CHECK (edition IN ('2014', '2024')),
    dndsu_id int NOT NULL,
    slug text NOT NULL,
    name_ru text NOT NULL DEFAULT '',
    name_en text NOT NULL DEFAULT '',
    cr text NOT NULL DEFAULT '',
    type text NOT NULL DEFAULT '',
    url text NOT NULL,
    is_unique_npc boolean NOT NULL DEFAULT false,
    content_hash text NOT NULL DEFAULT '',
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (section, edition, dndsu_id)
);

CREATE TABLE created_creatures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ru text NOT NULL DEFAULT '',
    name_en text NOT NULL DEFAULT '',
    statblock jsonb NOT NULL DEFAULT '{}'::jsonb,
    notes text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE player_characters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    ac int,
    passive_perception int,
    max_hp int,
    notes text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE encounters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    round int NOT NULL DEFAULT 1,
    active_combatant_id uuid,
    status text NOT NULL DEFAULT 'building',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE combatants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id uuid NOT NULL REFERENCES encounters(id) ON DELETE CASCADE,
    source_type text NOT NULL CHECK (source_type IN ('monster', 'created_creature', 'player_character', 'custom')),
    source_id uuid,
    monster_edition text,
    display_name text NOT NULL,
    max_hp int,
    current_hp int,
    temp_hp int NOT NULL DEFAULT 0,
    initiative int,
    conditions text[] NOT NULL DEFAULT '{}',
    notes text NOT NULL DEFAULT '',
    is_pc boolean NOT NULL DEFAULT false,
    sort_order int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_combatants_encounter ON combatants (encounter_id);

CREATE TABLE crawl_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind text NOT NULL DEFAULT 'incremental',
    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    status text NOT NULL DEFAULT 'running',
    added int NOT NULL DEFAULT 0,
    updated int NOT NULL DEFAULT 0,
    errors jsonb NOT NULL DEFAULT '[]'::jsonb
);
