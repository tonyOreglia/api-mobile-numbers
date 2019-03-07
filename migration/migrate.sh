#!/bin/sh
set -e

flyway -user="olx" -password="olx" -url="jdbc:postgresql://localhost/mobile_numbers" -locations=filesystem:migration/sql/ -table=flyway_schema_history migrate info || true
