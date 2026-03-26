#!/bin/sh

go build --tags fts5 -o gosimple
rm example.db
./gosimple
