CREATE TABLE IF NOT EXISTS Events ( 
    Id TEXT, 
    Title TEXT, 
    Author TEXT, 
    Date TIMESTAMP, 
    PRIMARY KEY (Id) 
);

INSERT INTO Events (Id, Title, Author, Date)
VALUES ('b6082a86-584a-40e7-9401-3e81945ef358', 'Some cool event', 'John Doe', '2024-01-01 00:00:00');