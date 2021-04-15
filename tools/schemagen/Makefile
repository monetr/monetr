
initial:
	go run github.com/monetrapp/rest-api/tools/schemagen > schema/00000000_Initial.up.sql
	yarn sql-formatter -l postgresql -u --lines-between-queries 2 schema/00000000_Initial.up.sql -o schema/00000000_Initial.up.sql


