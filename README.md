##pgxsqlizer

Generate go methods from your sql files.

args:

'-input' - folder with sql files
'-output' - folder where you want place generated files 

##Example: 
```go run main.go -input ./example/model -output ./example/model```

Result of generator will be in ```output/pgxsqlizer/*```