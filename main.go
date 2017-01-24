package main

import (
  "fmt"
  //"time"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

func main() {
  db, _ := sql.Open("mysql", "root:root@/a-ads_development")
  defer db.Close()

  err_conn := db.Ping()
  if err_conn != nil {
    fmt.Printf("err is %s\n", err_conn)
    return
  }
  rows, err := db.Query(query_pages("ranks_updated_at"))
  if err != nil {
    fmt.Printf("err is %s\n", err)
    return
  }
  defer rows.Close()

  var (
    id int
    ranks_updated_at sql.NullString
    ranks_string string
  )

  for rows.Next() {
    err := rows.Scan(&id, &ranks_updated_at)
    if err != nil {
      fmt.Printf("err is %s\n", err)
    }
    if ranks_updated_at.Valid {
      ranks_string = ranks_updated_at.String
      fmt.Println(id, ranks_string)
    } else {
      fmt.Printf("%s, %#v \n", id, ranks_updated_at)
    }
  }
}

func query_pages(order_by string) string {
  return fmt.Sprintf("SELECT id, ranks_updated_at FROM pages ORDER BY %s ASC", order_by)
}
