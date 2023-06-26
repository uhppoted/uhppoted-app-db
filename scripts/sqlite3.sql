CREATE TABLE ACL (
    Name       TEXT    DEFAULT '',
    CardNumber INTEGER UNIQUE,
    PIN        INTEGER DEFAULT 0,
    StartDate  TEXT    DEFAULT '',
    EndDate    TEXT    DEFAULT '',
    GreatHall  INTEGER DEFAULT 0,
    Gryffindor INTEGER DEFAULT 0,
    HufflePuff INTEGER DEFAULT 0,
    Ravenclaw  INTEGER DEFAULT 0,
    Slytherin  INTEGER DEFAULT 0,
    Kitchen    INTEGER DEFAULT 0,
    Dungeon    INTEGER DEFAULT 0,
    Hogsmeade  INTEGER DEFAULT 0
);

CREATE TABLE Events (
    Controller   INTEGER  NOT NULL,
    EventIndex   INTEGER  NOT NULL,
    Timestamp    DATETIME NULL,
    Type         INTEGER  NULL,
    Granted      INTEGER  NULL,
    Door         INTEGER  NULL,
    Direction    INTEGER  NULL,
    CardNumber   INTEGER  NULL,
    Reason       INTEGER  NULL,
    CONSTRAINT ControllerEventIndex UNIQUE (Controller, EventIndex) ON CONFLICT REPLACE
);

CREATE TABLE Audit (
    Timestamp  DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
    Operation  TEXT     DEFAULT '',
    Controller INTEGER  DEFAULT 0,
    CardNumber INTEGER  DEFAULT 0,
    Status     TEXT     DEFAULT '',
    Card       TEXT     DEFAULT ''
);

CREATE TABLE OperationsLog (
    Timestamp  DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
    Operation  TEXT     DEFAULT '',
    Controller INTEGER  NULL,
    Detail     TEXT     DEFAULT ''
);

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
