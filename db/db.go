package db

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "postgres"
)

type ItemsJSON struct {
	Chrt_id      int64  `json:"chrt_id"`
	Track_number string `json:"track_number"`
	Price        int64  `json:"price"`
	Rid          string `json:"rid"`
	Name         string `json:"name"`
	Sale         int64  `json:"sale"`
	Size         string `json:"size"`
	Total_price  int64  `json:"total_price"`
	Nm_id        int64  `json:"nm_id"`
	Brand        string `json:"brand"`
	Status       int64  `json:"status"`
}
type OrderJSON struct {
	Order_uid    string `json:"order_uid"`
	Track_number string `json:"track_number"`
	Entry        string `json:"entry"`

	Delivery struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction   string `json:"transaction"`
		Request_id    string `json:"request_id"`
		Currency      string `json:"currency"`
		Provider      string `json:"provider"`
		Amount        int64  `json:"amount"`
		Payment_dt    int64  `json:"payment_dt"`
		Bank          string `json:"bank"`
		Delivery_cost int64  `json:"delivery_cost"`
		Goods_total   int64  `json:"goods_total"`
		Custom_fee    int64  `json:"custom_fee"`
	} `json:"payment"`
	Items              []ItemsJSON `json:"items"`
	Locale             string      `json:"locale"`
	Internal_signature string      `json:"internal_signature"`
	Customer_id        string      `json:"customer_id"`
	Delivery_service   string      `json:"delivery_service"`
	Shardkey           string      `json:"shardkey"`
	Sm_id              int64       `json:"sm_id"`
	Date_created       string      `json:"date_created"`
	Oof_shard          string      `json:"oof_shard"`
}

func FillInMemoryDb() []OrderJSON {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)
	//main info about order
	rows, err := db.Query(fmt.Sprintf(`SELECT order_uid FROM public."order"`))
	CheckError(err)
	defer rows.Close()
	var bd_rows []OrderJSON
	for rows.Next() {
		var order_to_append OrderJSON
		err = rows.Scan(&order_to_append.Order_uid)
		order_to_append.SelectById(&order_to_append.Order_uid)
		bd_rows = append(bd_rows, order_to_append)
		CheckError(err)
	}
	return bd_rows
}

func (d *OrderJSON) PrintRowIntoWriter(w *http.ResponseWriter) {
	if d.Customer_id == "" {
		fmt.Fprintf(*w, "Wrong order id or its not set")
	} else {
		fmt.Fprintf(*w, "order_uid= %s\n track_numver= %s\n entry=%s\n delivery_name=%s\n delivery_phone=%s\n",
			d.Order_uid, d.Track_number, d.Entry, d.Delivery.Name, d.Delivery.Phone)

		fmt.Fprintf(*w, "delivery_zip=%s\n delivery_city=%s\n delivery_address=%s\n delivery_region=%s\n delivery_email=%s\n",
			d.Delivery.Zip, d.Delivery.City, d.Delivery.Address, d.Delivery.Region, d.Delivery.Region)

		fmt.Fprintf(*w, "payment_transaction_id=%s\n locale=%s\n customer_id=%s\n delivery_service=%s\n",
			d.Payment.Transaction, d.Locale, d.Customer_id, d.Delivery_service)

		fmt.Fprintf(*w, "shardkey=%s\n date_created=%s\n oof_shard=%s\n",
			d.Shardkey, d.Date_created, d.Oof_shard)

	}
}
func (d *OrderJSON) InsertInId() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)
	//items
	for i, elem := range d.Items {
		_, err = db.Exec("INSERT INTO items (chrt_id, track_number, price,rid,name,sale,size,total_price,nm_id,status,brand) VALUES ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10,$11)",
			elem.Chrt_id, elem.Track_number, elem.Price, elem.Rid, elem.Name, elem.Sale, elem.Size, elem.Total_price, elem.Nm_id, elem.Status, elem.Brand)
		if err != nil {
			fmt.Printf("Error while inserting into items %v, items num %d\n", err, i)
		}
	}
	//payments
	_, err = db.Exec("INSERT INTO payments (transaction, request_id, currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee) VALUES ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10)",
		d.Payment.Transaction, d.Payment.Request_id, d.Payment.Currency, d.Payment.Provider, d.Payment.Amount, d.Payment.Payment_dt, d.Payment.Bank,
		d.Payment.Delivery_cost, d.Payment.Goods_total, d.Payment.Custom_fee)
	if err != nil {
		fmt.Printf("Error while inserting into payments %v\n", err)
	}
	//order
	_, err = db.Exec("INSERT INTO public.\"order\" (order_uid, track_number, entry,delivery_name,delivery_phone,delivery_zip,delivery_city,delivery_address,delivery_region,delivery_email,payment_transaction_id,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard) VALUES ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)",
		d.Order_uid, d.Track_number, d.Entry, d.Delivery.Name, d.Delivery.Phone, d.Delivery.Zip, d.Delivery.City, d.Delivery.Address,
		d.Delivery.Region, d.Delivery.Email, d.Payment.Transaction, d.Locale, d.Internal_signature, d.Customer_id, d.Delivery_service, d.Shardkey, d.Sm_id, d.Date_created, d.Oof_shard)
	if err != nil {
		fmt.Printf("Error while inserting into order %v\n", err)
	}
	//order_items
	for i, elem := range d.Items {
		_, err = db.Exec("INSERT INTO order_items (order_uid, item_chrt_id) VALUES ($1,$2)", d.Order_uid, elem.Chrt_id)
		if err != nil {
			fmt.Printf("Error while inserting into order_items %v items num %d\n", err, i)
		}
	}
}
func (d *OrderJSON) SelectById(id *string) {

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)
	//main info about order
	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM public."order" Where order_uid='%s'`, *id))
	CheckError(err)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&d.Order_uid, &d.Track_number, &d.Entry, &d.Delivery.Name,
			&d.Delivery.Phone, &d.Delivery.Zip, &d.Delivery.City, &d.Delivery.Address,
			&d.Delivery.Region, &d.Delivery.Email, &d.Payment.Transaction,
			&d.Locale, &d.Internal_signature, &d.Customer_id,
			&d.Delivery_service, &d.Shardkey, &d.Sm_id, &d.Date_created, &d.Oof_shard)
		CheckError(err)
	}
	//payment info
	rows, err = db.Query(fmt.Sprintf(`SELECT * FROM public."payments" Where transaction='%s'`, d.Payment.Transaction))
	CheckError(err)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&d.Payment.Transaction, &d.Payment.Request_id, &d.Payment.Currency, &d.Payment.Provider,
			&d.Payment.Amount, &d.Payment.Payment_dt, &d.Payment.Bank, &d.Payment.Delivery_cost, &d.Payment.Goods_total, &d.Payment.Custom_fee)
		CheckError(err)
	}

	//from order_items to define items
	rows, err = db.Query(fmt.Sprintf(`SELECT * FROM public."order_items" Where order_uid='%s'`, d.Order_uid))
	CheckError(err)
	defer rows.Close()
	var nums []int64
	for rows.Next() {
		var new_num int64
		var id int
		err = rows.Scan(&id, &d.Order_uid, &new_num)
		nums = append(nums, new_num)
		CheckError(err)
	}
	//add items
	for num := range nums {
		rows, err = db.Query(fmt.Sprintf(`SELECT * FROM public."items" Where chrt_id='%d'`, num))
		CheckError(err)
		defer rows.Close()
		for rows.Next() {
			var new_item ItemsJSON
			err = rows.Scan(&new_item.Chrt_id, &new_item.Track_number, &new_item.Price, &new_item.Rid, &new_item.Name,
				&new_item.Sale, &new_item.Size, &new_item.Total_price, &new_item.Nm_id, &new_item.Brand, &new_item.Status)
			d.Items = append(d.Items, new_item)
			CheckError(err)
		}
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
