package pgxsqlizer

import "github.com/jackc/pgx/v5"

var userStmtMap = map[string]string{
	"GetUsers":   "select * from users where id = @userID and age > @userAge;",
	"InsertUser": "insert into users( id, name, age, email ) values ( @userID, @userName, @userAge, @userEmail );",
}

func GetUsers(userID string, userAge int) (string, pgx.NamesArgs) {
	return userStmtMap["GetUsers"], pgx.NamedArgs{
		"userID":  userID,
		"userAge": userAge,
	}
}

func InsertUser(userID string, userName string, userAge int, userEmail string) (string, pgx.NamesArgs) {
	return userStmtMap["InsertUser"], pgx.NamedArgs{
		"userID":    userID,
		"userName":  userName,
		"userAge":   userAge,
		"userEmail": userEmail,
	}
}
