package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"./db"

	"net/http"

	"log"
	"os"

	"github.com/nats-io/nats.go"
)

func FormHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	id := r.FormValue("name")
	var row_in_db db.OrderJSON
	row_in_db.SelectById(&id)
	row_in_db.PrintRowIntoWriter(&w)
}

func ValidateJSON(jsonData *[]byte, struct_to_write *db.OrderJSON) error {
	if err := json.Unmarshal(*jsonData, struct_to_write); err != nil {
		return err
	}
	//whole order
	if struct_to_write.Order_uid == "" || struct_to_write.Track_number == "" || struct_to_write.Entry == "" ||
		struct_to_write.Locale == "" || struct_to_write.Customer_id == "" || struct_to_write.Delivery_service == "" ||
		struct_to_write.Shardkey == "" || struct_to_write.Date_created == "" || struct_to_write.Oof_shard == "" {
		return errors.New("Error in order fields")
	}
	//payments
	if struct_to_write.Payment.Transaction == "" || struct_to_write.Payment.Currency == "" || struct_to_write.Payment.Provider == "" ||
		struct_to_write.Payment.Bank == "" || struct_to_write.Payment.Amount == 0 || struct_to_write.Payment.Payment_dt == 0 ||
		struct_to_write.Payment.Goods_total == 0 {
		return errors.New("Error in payment fields")
	}
	//delivery
	if struct_to_write.Delivery.Name == "" || struct_to_write.Delivery.Phone == "" || struct_to_write.Delivery.Zip == "" ||
		struct_to_write.Delivery.City == "" || struct_to_write.Delivery.Address == "" || struct_to_write.Delivery.Region == "" ||
		struct_to_write.Delivery.Email == "" {
		return errors.New("Error in delivery fields")
	}
	//items
	for _, elem := range struct_to_write.Items {
		if elem.Brand == "" || elem.Name == "" || elem.Rid == "" || elem.Size == "" || elem.Track_number == "" || elem.Chrt_id == 0 || elem.Nm_id == 0 ||
			elem.Price == 0 || elem.Sale == 0 || elem.Status == 0 || elem.Total_price == 0 {
			return errors.New("Error in item fields")
		}
	}
	return nil
}

func main() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", FormHandler)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf("Error connect %v", err)
	}

	bd_rows := db.FillInMemoryDb()
	defer nc.Close()
	nc.Subscribe("foo", func(m *nats.Msg) {
		var somevar db.OrderJSON
		error_code := ValidateJSON(&m.Data, &somevar)
		if error_code != nil {
			fmt.Printf("Error while validation %v\n", error_code)
		} else {
			if err := os.WriteFile("reserve_copy.txt", []byte(m.Data), 0666); err != nil {
				log.Fatal(err)
			}
			//panic("123")
			somevar.InsertInId()
			bd_rows = append(bd_rows, somevar)
			err := os.Remove("reserve_copy.txt")
			if err != nil {
				log.Fatalf("Error removing file: %v", err)
			}
		}
	})
	b, err := os.ReadFile("reserve_copy.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	} else {
		nc.Publish("foo", b)
	}
	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
