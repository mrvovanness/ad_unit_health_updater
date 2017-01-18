package main

import (
  "fmt"
  "time"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

func main() {
  start_time := time.Now()
  fmt.Printf("The time now is %s\n", start_time)
  db, err := sql.Open("mysql", "dev:1111@/a-ads_development")
  fmt.Printf("db is %s\n, error is %s\n", db, err)
  defer db.Close()

  err2 := db.Ping()
  if err2 != nil {
    fmt.Printf("err is %s\n", err2)
  } else {
    fmt.Println("Connection working")
  }
}
