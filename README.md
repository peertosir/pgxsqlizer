# sql2gogen

Generate go methods from your sql files.

### Possible arguments:
- ```-input <folder>``` - folder with sql files
- ```-output <folder>``` - folder where you want to place generated files
- ```-genPkg <pkgname>``` - package + folder name for generated files. Default is 'actionsgen'
- ```-returnType <pkgname>``` - package + folder name for generated files. Default is 'slice'
- ```-placeholder <placeholder>``` - placeholder which will be used in queries. Options: [@|?|$]. 
```
select * from users where id = $1
select * from users where id = @userID
select * from users where id = ?
```
'?' - for generic usage (usually not with psql)
'$<int>' - generic psql placeholder
'@<string>' - pgx/v5 named args compatibility. pgx.NamedArgs is an alias for map[string]any 

### Example: 
```go run . -input ./example/model -output ./example/model -returnType map -placeholder @ -genPkg generatedactions```

Result of generator will be in ```./example/model/generatedactions/*```

### Rules for SQL Files.

1. EVERY QUERY MUST HAVE A NAME. To name your query enter just before your query:
```-- title: <somename>```
2. Describe your params in query with pattern:
```@<paramName>:<paramGoType>@```
3. If your go type should be imported in generated file, please add ```-- addimport: "<pkg>"``` in any place in your sql file. Import comment works only for 1 file

Example of sqls you can find in example folder.
