CREATE TABLE IF NOT EXISTS waypoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION NOT NULL CHECK (latitude >= -90 AND latitude <= 90),
    longitude DOUBLE PRECISION NOT NULL CHECK (longitude >= -180 AND longitude <= 180),
    "order" INTEGER NOT NULL CHECK ("order" >= 0),
    name VARCHAR(200),
    address VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(route_id, "order")
);

CREATE INDEX idx_waypoints_route_id ON waypoints(route_id);
CREATE INDEX idx_waypoints_order ON waypoints(route_id, "order");
