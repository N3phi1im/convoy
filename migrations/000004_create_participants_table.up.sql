CREATE TABLE IF NOT EXISTS participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'left', 'removed')),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    UNIQUE(route_id, user_id)
);

CREATE INDEX idx_participants_route_id ON participants(route_id);
CREATE INDEX idx_participants_user_id ON participants(user_id);
CREATE INDEX idx_participants_status ON participants(status);
