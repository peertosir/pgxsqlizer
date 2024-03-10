-- title: GetUsers
select * from users
where id = @userID:string@ and age > @userAge:int@;

-- title: InsertUser
insert into users(
    id,
    name,
    age,
    email
) values (
    @userID:string@,
    @userName:string@,
    @userAge:int@,
    @userEmail:string@
);