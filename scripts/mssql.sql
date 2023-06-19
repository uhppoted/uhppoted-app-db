CREATE TABLE ACL (
    Name       VARCHAR(256) DEFAULT '',
    CardNumber INT     UNIQUE,
    PIN        INT     DEFAULT 0,
    StartDate  DATE    NULL,
    EndDate    DATE    NULL,
    GreatHall  TINYINT DEFAULT 0,
    Gryffindor TINYINT DEFAULT 0,
    HufflePuff TINYINT DEFAULT 0,
    Ravenclaw  TINYINT DEFAULT 0,
    Slytherin  TINYINT DEFAULT 0,
    Kitchen    TINYINT DEFAULT 0,
    Dungeon    TINYINT DEFAULT 0,
    Hogsmeade  TINYINT DEFAULT 0
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
