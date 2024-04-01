#!/bin/bash

curl -X GET -H "Content-Type: application/json" \
"http://127.0.0.1:8080/api/v1/ad?offset=1&limit=1&age=25&gender=F&country=TW&platform=ios"

echo