// Copyright 2023 uhppoted@twyst.co.za. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

/*
Package uhppoted-app-db integrates the uhppote-core API with access control lists stored in a database.

uhppoted-app-db can be used from the command line but is really intended to be run from a cron job to maintain
the cards and permissions on a set of access controllers from a unified access control list (ACL) stored in a
database.

uhppoted-app-db supports the following commands:

  - get-acl, to retrieve an ACL from a database and store it in a file
  - put-acl, to extract an ACL from a file and store it to a database
  - load-acl, to download an ACL from a database to a set of access controllers
  - store-acl, to retrieve the ACL from a set of controllers and store it in a database table
  - compare-acl, to compare an ACL from a database table with the cards and permissons on a set of access controllers
*/
package db
