// Code generated by sql2gogen. DO NOT EDIT.
package actionsgen
	
const getUsers = "select * from users where id = @userID and age > @userAge;"	
const insertUser = "insert into users( id, name, age, email ) values ( @userID, @userName, @userAge, @userEmail );"

func GetUsers(
	userID string, userAge int,
) (string, map[string]any) {
	return getUsers, map[string]any{
		"userID": userID, "userAge": userAge,
	}
}

func InsertUser(
	userID string, userName string, userAge int, userEmail string,
) (string, map[string]any) {
	return insertUser, map[string]any{
		"userID": userID, "userName": userName, "userAge": userAge, "userEmail": userEmail,
	}
}

