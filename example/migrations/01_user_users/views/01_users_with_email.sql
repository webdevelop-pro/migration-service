create view public.users_with_email as
select * from user_users where email is not null;