DROP TABLE IF EXISTS salary_history;
DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS employee_emergency_contact;
DROP TABLE IF EXISTS employee_family;
ALTER TABLE employees
    DROP COLUMN IF EXISTS cv_url,
    DROP COLUMN IF EXISTS contract_end_date;
