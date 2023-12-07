-- this is only necessary for the local and localdev environments
-- kubectl port-forward svc/postgres-postgresql-primary 5436:5432 -n postgres
-- (from root) make db-pg-init
-- user creation
CREATE USER APP_NAME_UND_app WITH PASSWORD 'DB_PASSWORD';

CREATE DATABASE APP_NAME_UND_local;

CREATE DATABASE APP_NAME_UND_localdev;

GRANT ALL PRIVILEGES ON DATABASE APP_NAME_UND_local TO APP_NAME_UND_app;

GRANT ALL PRIVILEGES ON DATABASE APP_NAME_UND_localdev TO APP_NAME_UND_app;
