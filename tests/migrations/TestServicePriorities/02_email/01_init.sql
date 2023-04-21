
CREATE TABLE email_emails (
      id serial primary key,
      user_id integer not null,
      FOREIGN KEY (user_id) REFERENCES user_users ("id")
          ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED);
CREATE INDEX email_emails_user_fk ON email_emails USING btree ("user_id");