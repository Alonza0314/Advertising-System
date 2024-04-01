#!/bin/bash

# start the mongodb service
systemctl start mongod

# set the configuration file path
source project.conf

# if database does not exist, create it; otherwis, drop it and re-create again
# and make a collection to store the ads
mongosh "$uri" --eval "db = db.getSiblingDB('$database'); if (db.getCollectionNames().length > 0) { db.dropDatabase(); } db.createCollection('$collection'); quit();"

# tidy the module have been used
go mod tidy

# complete
echo "Finish"
