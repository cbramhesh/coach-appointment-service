DROP TABLE IF EXISTS bookings;

CREATE TABLE bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    coach_id INT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active_slot VARCHAR(100) GENERATED ALWAYS AS (
        CASE WHEN status = 'confirmed'
             THEN CONCAT(coach_id, '_', start_time)
             ELSE NULL
        END
    ) STORED,
    UNIQUE INDEX idx_no_double_booking (active_slot)
);

CREATE INDEX idx_bookings_user ON bookings (user_id, start_time);
CREATE INDEX idx_bookings_coach_time ON bookings (coach_id, start_time);
