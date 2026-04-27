CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    route_type VARCHAR(50) NOT NULL CHECK (route_type IN ('driving', 'cycling', 'walking', 'running')),
    visibility VARCHAR(50) NOT NULL CHECK (visibility IN ('public', 'private', 'invite_only')),
    start_time TIMESTAMP,
    max_participants INTEGER CHECK (max_participants > 0),
    distance DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0,
    difficulty VARCHAR(50),
    geometry TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'in_progress', 'completed', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_routes_creator_id ON routes(creator_id);
CREATE INDEX idx_routes_status ON routes(status);
CREATE INDEX idx_routes_visibility ON routes(visibility);
CREATE INDEX idx_routes_start_time ON routes(start_time);
CREATE INDEX idx_routes_route_type ON routes(route_type);
CREATE INDEX idx_routes_created_at ON routes(created_at);
CREATE INDEX idx_routes_deleted_at ON routes(deleted_at);