package main

import (
  "fmt"
  "time"
  "database/sql"
  "os"
  "sync"
  "sync/atomic"
  "net"
  "net/http"
  "io/ioutil"
  "regexp"
  "math"
  _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
  db, _ = sql.Open("mysql", "root:root@/a-ads_development")
  check_conn()
}

type AdUnit struct {
  id int
  health float64
  health_updated_at sql.NullString
  site_url string
}

func (ad_unit AdUnit) String() string {
  return fmt.Sprintf("id: %d, health: %f, health_updated_at: %s, url: %s\n",
    ad_unit.id, ad_unit.health, ad_unit.health_updated_at.String, ad_unit.site_url)
}

func main() {
  defer db.Close()
  rows, err := db.Query(ad_unit_query(2000, day_ago()))
  var ad_units []*AdUnit

  if err != nil {
    fmt.Printf("err is %s\n", err)
    return
  }

  for rows.Next() {
    ad_unit := &AdUnit{}
    err := rows.Scan(&ad_unit.id, &ad_unit.health, &ad_unit.health_updated_at, &ad_unit.site_url)
    if err != nil { fmt.Printf("err is %s\n", err) }
    ad_units = append(ad_units, ad_unit)
  }
  rows.Close()

  var wait_group sync.WaitGroup
  var ops uint64 = 0
  wait_group.Add(len(ad_units))
  netTransport := &http.Transport{
    Dial: (&net.Dialer{
      Timeout: 10 * time.Second,
    }).Dial, TLSHandshakeTimeout: 10 * time.Second,
  }

  client := &http.Client{
    Timeout: time.Second * 10,
    Transport: netTransport,
  }

  for _, ad_unit := range ad_units {
    go func(ad_unit *AdUnit) {

      req, _ := http.NewRequest("GET", ad_unit.site_url, nil)
      req.Header.Add("Accept", "text/html")
      req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
      req.Header.Add("Connection", "close")
      res, err := client.Do(req)
      if err != nil {
        //fmt.Println(err)
        atomic.AddUint64(&ops, 1)
        wait_group.Done()
        return
      }
      html, _ := ioutil.ReadAll(res.Body)
      res.Body.Close()
      pattern := fmt.Sprintf("data-aa=.%d.", ad_unit.id)
      is_healthy, err := regexp.MatchString(pattern, string(html))
      speed := 0.25
      if is_healthy {
        health := math.Min(ad_unit.health * (1 - speed) + 1 * speed, 1.0)
        update_ad_unit(ad_unit.id, health)
      } else {
        health := math.Max(ad_unit.health * (1 - speed) + 0 * speed, 0)
        update_ad_unit(ad_unit.id, health)
      }
      fmt.Printf("result: %t, error: %s\n", is_healthy, err)
      wait_group.Done()
    }(ad_unit)
  }
  wait_group.Wait()

  ops_final := atomic.LoadUint64(&ops)
  fmt.Printf("not found: %d\n", ops_final)
}

func ad_unit_query(batch_size int, from_time string) string {
  query := "SELECT id, health, health_updated_at, site_url FROM ad_units " +
           "WHERE (deleted_at IS NULL) " +
           "AND ad_unit_type = 'site' " +
           "AND ((health > 0.0 OR weight > 0.0) OR created_at > '%s') " +
           "ORDER BY health_updated_at ASC LIMIT %d"
  return fmt.Sprintf(query, from_time, batch_size)
}

func update_ad_unit(id int, health float64) {
  query := fmt.Sprintf(
    "UPDATE ad_units SET health = %f, health_updated_at = NOW() WHERE ad_units.id = %d;",
    health, id)
  db.Exec(query)
}

func day_ago() string {
  now := time.Now().UTC()
  day_ago := now.AddDate(0, 0, -1)
  formatted := fmt.Sprintf( "%d-%02d-%02d %02d:%02d:%02d",
    day_ago.Year(), day_ago.Month(), day_ago.Day(),
    day_ago.Hour(), day_ago.Minute(), day_ago.Second())
  return formatted
}

func check_conn() {
  err := db.Ping()
  if err != nil {
    fmt.Printf("%s, exiting\n", err)
    os.Exit(1)
  }
}
