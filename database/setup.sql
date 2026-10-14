IF DB_ID(N'LearningStack') IS NULL
BEGIN
    CREATE DATABASE LearningStack;
END;
GO

USE LearningStack;
GO

IF OBJECT_ID(N'dbo.Customers', N'U') IS NULL
BEGIN
    CREATE TABLE dbo.Customers
    (
        CustomerId INT NOT NULL CONSTRAINT PK_Customers PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL,
        Email NVARCHAR(255) NULL
    );
END;
GO

IF NOT EXISTS (SELECT 1 FROM dbo.Customers WHERE CustomerId = 101)
BEGIN
    INSERT INTO dbo.Customers (CustomerId, Name, Email)
    VALUES (101, N'Rahul Sharma', N'rahul@example.com');
END;
GO

CREATE OR ALTER PROCEDURE dbo.GetCustomerById
    @CustomerId INT
AS
BEGIN
    SET NOCOUNT ON;

    SELECT CustomerId, Name, Email
    FROM dbo.Customers
    WHERE CustomerId = @CustomerId;
END;
GO
