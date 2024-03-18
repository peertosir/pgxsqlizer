-- title: GetProjects
select * from projects
where id = @projectID:string@;

-- title: InsertProjects
-- addimport: "time"
insert into projects(
    id,
    title,
    description,
    createdAt
) values (
    @projectID:string@,
    @projectTitle:string@,
    @projectDescription:string@,
    @createdAt:time.Duration@
) where projectId=@projectID:string@;
