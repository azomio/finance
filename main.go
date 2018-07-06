package main

import (
    "html/template"
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

type Receipt struct {
	Fp string
	I string
	Fn string
	Sum string
	AddTime int
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

type ReceiptsPage struct {
    Receipts []Receipt
}

func main() {
	fmt.Println("Started")

	http.HandleFunc("/receipt/add", receiptAdd)
	http.HandleFunc("/receipt/delete", receiptDelete)
	http.HandleFunc("/receipt/fetch", receiptFetch)
    http.HandleFunc("/", mainPageHandler)
    http.HandleFunc("/get", addReceiptHandler)
    http.HandleFunc("/goods", goodsListHandler)
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err);
    }

}

func receiptFetch (w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		fetchData := make(map[string]string)
	    fetchData["fn"] = r.Form["fn"][0]
	    fetchData["fp"] = r.Form["fp"][0]
	    fetchData["i"] = r.Form["i"][0]

   		// db, err := sql.Open("sqlite3", "./foo.db")
     //    checkErr(err)

	    parsed := fetchReceipt(fetchData)
	    fmt.Println(parsed)
	}
}

func receiptDelete (w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
    	fn := r.Form["fn"][0]
    	fp := r.Form["fp"][0]
    	i := r.Form["i"][0]

   		db, err := sql.Open("sqlite3", "./foo.db")
        checkErr(err)

		stmt, err := db.Prepare("DELETE FROM receipt WHERE fn = ? AND fp = ? AND i = ?")
		checkErr(err)
		_, err = stmt.Exec(fn, fp, i)
		checkErr(err)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func receiptAdd (w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		return
	}

	r.ParseForm()

	var fn, fp, i string

	if query, ok := r.Form["query"]; ok && len(query[0]) != 0 {
		query := r.Form["query"][0]
		data, err := url.ParseQuery(query)
		fmt.Println(data, err)
		fn = data["fn"][0]
		fp = data["fp"][0]
		i = data["i"][0]
	} else {
		fn = r.Form["fn"][0]
		fp = r.Form["fp"][0]
		i = r.Form["i"][0]
	}


	db, err := sql.Open("sqlite3", "./foo.db")
    checkErr(err)

	stmt, err := db.Prepare("INSERT INTO receipt(fn, fp, i) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(fn, fp, i)
	fmt.Println(err)

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}

// t=20170926T2012&s=507.00&fn=8710000100993415&i=7269&fp=3426724739&n=1
func mainPageHandler (w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "./foo.db")
    checkErr(err)

    rows, err := db.Query("SELECT * FROM receipt")
    checkErr(err)

    var fn string
    var i string
    var fp string
    var sum string
    var data string
    var addTime int

	var receipts []Receipt
    for rows.Next() {
    	err = rows.Scan(&fn, &i, &fp, &sum, &data, &addTime)
		receipts = append(receipts, Receipt{
			Fn: fn,
			Fp: fp,
			I: i,
			Sum: sum,
			AddTime: addTime,
		})
    }
    rows.Close()

	t, err := template.ParseFiles("tmpl/receipts.html")
    checkErr(err)

    page := ReceiptsPage{receipts}
    t.Execute(w, page)
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

    t, err := template.ParseFiles("tmpl/goods_list.html")
    checkErr(err)

    page := GoodsPage{goods}
    t.Execute(w, page)
}

func addReceiptHandler (w http.ResponseWriter, r *http.Request) {

    r.ParseForm()
    fmt.Println(r.Form)

    t, err := template.ParseFiles("tmpl/index.html")
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

}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func fetchReceipt(data map[string]string) ReceiptResp {

	client := &http.Client{}

    url := "http://proverkacheka.nalog.ru:8888/v1/inns/*/kkts/*/fss/" + data["fn"] + "/tickets/" + data["i"] + "?fiscalSign=" + data["fp"] + "&sendToEmail=no"
    fmt.Println(url)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Basic Kzc5MTU0NTQ1MDExOjE1NDg4NQ==")
    req.Header.Set("Device-Id", "7a41e33b11e458c")
    req.Header.Set("Device-OS", "Android 6.0")
    resp, err := client.Do(req)
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    log.Print("before")
    log.Print(body)
    var parsed ReceiptResp
    if (err != nil) {
        fmt.Printf("Error: %s", err);
    } else {
        _ = json.Unmarshal(body, &parsed)
    }

    return parsed
}

func check_receipt(input string) ReceiptResp {

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

    fetchData := make(map[string]string)
    fetchData["fn"] = data["fn"][0]
    fetchData["fp"] = data["fp"][0]
    fetchData["i"] = data["i"][0]

    parsed := fetchReceipt(fetchData)
 
    return parsed
}
