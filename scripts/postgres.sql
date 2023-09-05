CREATE DATABASE uhppoted;
USE uhppoted;

CREATE TABLE ACLx (
    Name       VARCHAR(256) DEFAULT '',
    CardNumber INT      UNIQUE,
    PIN        INT      DEFAULT 0,
    StartDate  DATE     NULL,
    EndDate    DATE     NULL,
    GreatHall  SMALLINT DEFAULT 0,
    Gryffindor SMALLINT DEFAULT 0,
    HufflePuff SMALLINT DEFAULT 0,
    Ravenclaw  SMALLINT DEFAULT 0,
    Slytherin  SMALLINT DEFAULT 0,
    Kitchen    SMALLINT DEFAULT 0,
    Dungeon    SMALLINT DEFAULT 0,
    Hogsmeade  SMALLINT DEFAULT 0
);

CREATE TABLE Audit (
    Timestamp  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    Operation  VARCHAR(64)  DEFAULT '',
    Controller INT          DEFAULT 0,
    CardNumber INT          DEFAULT 0,
    Status     VARCHAR(64)  DEFAULT '',
    Card       VARCHAR(255) DEFAULT ''
);

CREATE TABLE OperationsLog (
    Timestamp  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    Operation  VARCHAR(64)  DEFAULT '',
    Controller INT          NULL,
    Detail     VARCHAR(255) DEFAULT ''
);


CREATE USER uhppoted PASSWORD 'qwerty';

GRANT SELECT,INSERT,UPDATE,DELETE ON ACL           TO uhppoted;
GRANT SELECT,INSERT,UPDATE,DELETE ON Audit         TO uhppoted;
GRANT SELECT,INSERT,UPDATE,DELETE ON OperationsLog TO uhppoted;

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Albus Dumbledore', 10058400, 0, '2023-01-01', '2023-12-31', 1,1,1,1,1,1,1,1);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Hagrid', 10058401, 0, '2023-01-01', '2023-12-31', 1,1,1,1,1,1,0,1);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Dobby', 10058402, 0, '2023-01-01', '2023-12-31', 1,1,1,1,1,1,0,1);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Harry Potter', 10058403, 0, '2023-01-01', '2023-12-31', 1,1,0,0,0,0,0,0);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Hermione Grainger', 10058404, 0, '2023-01-01', '2023-12-31', 1,1,0,0,0,0,1,0);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Crookshanks', 10058405, 0, '2023-01-01', '2023-12-31', 0,1,0,0,0,1,0,0);

INSERT INTO ACL    (Name, CardNumber,PIN,StartDate,EndDate,GreatHall,Gryffindor,HufflePuff,Ravenclaw,Slytherin,Kitchen,Dungeon,Hogsmeade)
            VALUES ('Tom Riddle', 10058406, 0, '2023-01-01', '2023-12-31', 0,0,0,0,0,0,1,1);
