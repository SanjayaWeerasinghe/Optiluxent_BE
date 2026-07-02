-- The QCLine GORM entity has CreatedAt/UpdatedAt fields but the table is missing
-- the columns, breaking INSERTs from auto-QC creation.
ALTER TABLE quality_check_lines
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
