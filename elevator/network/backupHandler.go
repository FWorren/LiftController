package network

import (
	"../types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

func Write_to_file(client Client) {
	client_encoded, err_encoding := json.Marshal(client)
	if err_encoding != nil {
		fmt.Println("error encoding json: ", err_encoding)
	}

	filename := "network/backup/data"
	filename += Get_last_ip_digits(client.Ip)

	_, err := os.Open(filename)
	if err != nil {
		fmt.Println("No file to write to, creating file..")
		_, _ = os.Create(filename)
	}

	err = ioutil.WriteFile(filename, client_encoded, 0644)
	if err != nil {
		fmt.Println("error writing to file")
		log.Fatal(err)
	}
}

func Read_file(file string) (bool, Client) {
	var decoded_client Client
	file_opened, err := os.Open(file)
	if err != nil {
		fmt.Println("No file to read from, creating file..")
		_, _ = os.Create(file)
		return true, decoded_client
	}
	data := make([]byte, 1024)
	n, err1 := file_opened.Read(data)
	if err1 != nil {
		fmt.Println("error reading file")
		fmt.Println(err1)
	}
	err_decoding := json.Unmarshal(data[:n], &decoded_client)
	if err_decoding != nil {
		fmt.Println("error decoding client from backup file")
	}
	return false, decoded_client
}

func Restore_command_orders(check_backup_c chan Client, localIP net.IP) bool {
	ip_last_digits := Get_last_ip_digits(localIP)
	file := "network/backup/data"
	file += ip_last_digits
	err, backup_client := Read_file(file)
	if !err {
		check_backup_c <- backup_client
		return true
	}
	return false
}

func Get_last_ip_digits(ip net.IP) string {
	ipstring := strings.Split(ip.String(), ".")
	return ipstring[len(ipstring)-1]
}
