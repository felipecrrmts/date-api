db = db.getSiblingDB('date');
db.createCollection('users');
db.users.createIndex({ location: "2dsphere" });