--- update updated_at automatically on every update operation
CREATE OR REPLACE FUNCTION update_at_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--- automatically update updated_at field
DROP TRIGGER IF EXISTS set_timestamp_user_users ON user_users;
CREATE TRIGGER set_timestamp_user_users
  BEFORE UPDATE ON user_users
  FOR EACH ROW
  EXECUTE PROCEDURE update_at_set_timestamp();
COMMIT;
