#!/bin/sh

set -eu

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" <<-EOF
CREATE USER trustreadyd;
ALTER USER trustreadyd WITH SUPERUSER;
ALTER USER trustreadyd PASSWORD 'trustreadyd';
CREATE DATABASE trustreadyd;
GRANT ALL PRIVILEGES ON DATABASE trustreadyd TO trustreadyd;
CREATE DATABASE trustreadyd_test;
GRANT ALL PRIVILEGES ON DATABASE trustreadyd_test TO trustreadyd;
EOF

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d trustreadyd <<-EOF
ALTER SCHEMA public OWNER TO trustreadyd;
GRANT ALL ON SCHEMA public TO trustreadyd;
ALTER DATABASE trustreadyd SET probo.trust_center_base_domain TO 'probopage.localhost';
EOF

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d trustreadyd_test <<-EOF
ALTER SCHEMA public OWNER TO trustreadyd;
GRANT ALL ON SCHEMA public TO trustreadyd;
ALTER DATABASE trustreadyd_test SET probo.trust_center_base_domain TO 'probopage.localhost';
EOF
