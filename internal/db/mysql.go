package db


import (
"database/sql"
"fmt"
)


type MySQL struct {
DB *sql.DB
}


func NewMySQL(user, pass, host string, port int, database string) (*MySQL, error) {
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, pass, host, port, database)
db, err := sql.Open("mysql", dsn)
if err != nil {
return nil, err
}
return &MySQL{DB: db}, nil
}


func (m *MySQL) Close() error {
return m.DB.Close()
}