#!/bin/bash

curl -X POST -H "Content-Type: application/json" \
"http://127.0.0.1:8080/api/v1/ad" \
--data '{
"title": "AD 66",
"startAt": "2023-12-10T03:00:00.000Z",
"endAt": "2024-06-21T16:00:00.000Z",
"conditions": [
{
"ageStart": 20,
"ageEnd": 30,
"gender": ["F"],
"country": ["TW", "JP"],
"platform": ["android", "ios"]
}
]
}'

echo

curl -X POST -H "Content-Type: application/json" \
"http://127.0.0.1:8080/api/v1/ad" \
--data '{
"title": "AD 55",
"startAt": "2023-12-10T03:00:00.000Z",
"endAt": "2024-05-21T16:00:00.000Z",
"conditions": [
{
"ageStart": 20,
"ageEnd": 30,
"country": ["TW", "JP"],
"platform": ["android", "ios"]
}
]
}'

echo

curl -X POST -H "Content-Type: application/json" \
"http://127.0.0.1:8080/api/v1/ad" \
--data '{
"title": "AD 02",
"startAt": "2023-12-10T03:00:00.000Z",
"endAt": "2024-04-21T16:00:00.000Z",
"conditions": [
{
"ageStart": 20,
"ageEnd": 30,
"country": ["TW", "JP"],
"platform": ["android", "ios"]
}
]
}'

echo