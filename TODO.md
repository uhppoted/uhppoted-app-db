# TODO

## IN PROGRESS

- [x] Microsoft SQL server support (cf. https://github.com/uhppoted/uhppoted-app-db/issues/3)
- [x] Audit trail (cf. https://github.com/uhppoted/uhppoted-app-db/issues/4)
- [x] Operations log (cf. https://github.com/uhppoted/uhppoted-app-db/issues/4)
- [x] Events (cf. https://github.com/uhppoted/uhppoted-app-db/issues/5)
- [x] MySQL support (cf. https://github.com/uhppoted/uhppoted-app-db/issues/1)
      - [x] get-acl
      - [x] put-acl
      - [x] load-acl
      - [x] store-acl
      - [x] compare-acl
      - [x] get-events
      - [x] audit trail
      - [x] log
      - [x] DSN: host/port
      - [x] Rename StoreEvents to PutEvents
      - [x] README
      - [x] CHANGELOG
      - [x] record2row: use []uint8 for DATE fields

- [x] Clean up DSN/DB logic
      - [x] get-acl
      - [x] put-acl
      - [x] load-acl
      - [x] store-acl
      - [x] compare-acl
      - [x] get-events

- [ ] sqlite3: replace UPSERTs with TABLE ON CONFLICT REPLACE clause

## TODO

- [ ] Postgres
- [ ] Firebase
- [ ] DynamoDB
- [ ] [Xata](https://xata.io)
- [ ] [liveblocks](https://liveblocks.io)
- [ ] [MyCelial](https://github.com/mycelial)

## Notes

1. https://leerob.io/blog/backend