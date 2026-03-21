CREATE TABLE IF NOT EXISTS availabilities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    coach_id INT NOT NULL,
    day_of_week TINYINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_avail_coach FOREIGN KEY (coach_id) REFERENCES coaches(id) ON DELETE CASCADE,
    CONSTRAINT chk_valid_window CHECK (end_time > start_time),
    CONSTRAINT uq_coach_availability UNIQUE (coach_id, day_of_week, start_time, end_time)
);

CREATE INDEX idx_avail_coach_day ON availabilities (coach_id, day_of_week);
