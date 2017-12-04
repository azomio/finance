package main

import (
    "text/template"
    "fmt"
    "net/http"
    "net/url"
    "log"
    "io/ioutil"
    "encoding/json"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type ReceiptResp struct {
    Document struct {
        Receipt struct {
            Items []Item `json:"items"`
        } `json:"receipt"`
    } `json:"document"`
}

type Item struct {
    Sum int `json:"sum"`
    Name string `json:"name"`
    Quantity int `json:"quantity"`
    Price int `json:"price"`
}

type GoodsItem struct {
    Id int
    Descr string
    Price string
    Time int
}

type ItemView struct {
    Sum float64
    Name string
    Quantity int
    Price int
}

type Page struct {
    Code string
    Items []ItemView
}

type GoodsPage struct {
    Goods []GoodsItem
}

func main() {

    http.HandleFunc("/add", addReceiptHandler)
    http.HandleFunc("/", goodsListHandler)
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err);
    }

}

func goodsListHandler (w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open("sqlite3", "./foo.db")
    checkErr(err)

    rows, err := db.Query("SELECT * FROM goods")
    checkErr(err)
    var id int
    var descr string
    var price string
    var time int

    var goods []GoodsItem
    for rows.Next() {
        err = rows.Scan(&id, &descr, &price, &time)
        checkErr(err)
        goods = append(goods, GoodsItem{id, descr, price, time})
    }

    rows.Close() //good habit to close

    t, err := template.ParseFiles("goods_list.html")
    checkErr(err)

    page := GoodsPage{goods}
    t.Execute(w, page)
}

func addReceiptHandler (w http.ResponseWriter, r *http.Request) {

    r.ParseForm()
    fmt.Println(r.Form)

    t, err := template.ParseFiles("index.html")
    checkErr(err)

    var data ReceiptResp
    codes := r.Form["code"]
    page := Page{}
    if len(codes) > 0 {
        code := codes[0]
        page.Code = code
        data = check_receipt(code)
    }

    var goods []ItemView
    for _, item := range data.Document.Receipt.Items {
        real_sum := float64(item.Sum) / 100
        goods = append(goods, ItemView{ Sum: real_sum, Name: item.Name, Quantity: item.Quantity, Price: item.Price} )
    }
    page.Items = goods

    fmt.Println(page)
    t.Execute(w, page)

    // // insert
    // stmt, err := db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
    // checkErr(err)

    // res, err := stmt.Exec("astaxie", "研发部门", "2012-12-09")
    // checkErr(err)

    // id, err := res.LastInsertId()
    // checkErr(err)

    // fmt.Println(id)
    // // update
    // stmt, err = db.Prepare("update userinfo set username=? where uid=?")
    // checkErr(err)

    // res, err = stmt.Exec("astaxieupdate", id)
    // checkErr(err)

    // affect, err := res.RowsAffected()
    // checkErr(err)

    // fmt.Printf(w, affect)

    // // query


    // // delete
    // stmt, err = db.Prepare("delete from userinfo where uid=?")
    // checkErr(err)

    // res, err = stmt.Exec(id)
    // checkErr(err)

    // affect, err = res.RowsAffected()
    // checkErr(err)

    // fmt.Printf(w, affect)

    // db.Close()

}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func check_receipt(input string) ReceiptResp {

    client := &http.Client{}

    data, err := url.ParseQuery(input)
    if err != nil {
        fmt.Println(err)
        return ReceiptResp{}
    }
    fn := data["fn"]
    if len(fn) == 0 {
        return ReceiptResp{}
    }
    fp := data["fp"]
    if len(fp) == 0 {
        return ReceiptResp{}
    }
    i := data["i"]
    if len(i) == 0 {
        return ReceiptResp{}
    }

    url := "http://proverkacheka.nalog.ru:8888/v1/inns/*/kkts/*/fss/" + fn[0] + "/tickets/" + i[0] + "?fiscalSign=" + fp[0] + "&sendToEmail=no"
    fmt.Println(url)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Basic Kzc5MTU0NTQ1MDExOjE1NDg4NQ==")
    req.Header.Set("Device-Id", "7a41e33b11e458c")
    req.Header.Set("Device-OS", "Android 6.0")
    resp, err := client.Do(req)
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    var parsed ReceiptResp
    if (err != nil) {
        fmt.Printf("Error: %s", err);
    } else {
        _ = json.Unmarshal(body, &parsed)
    }
    return parsed
}