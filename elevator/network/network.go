package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func Inter_process_communication(msg_from_network chan Client, order_from_network chan Client, order_from_cost chan Client, lost_orders_c chan Client, set_light_c chan Lights, del_order_c chan Order, localIP net.IP, all_clients map[string]Client, order_complete_c chan Order) {
	for {
		select {
		case new_order := <-msg_from_network:
			all_clients[new_order.Ip.String()] = new_order
			if new_order.Button != BUTTON_COMMAND {
				network_list[new_order.Button][new_order.Floor] = true
				priorityHandler(new_order, order_from_cost, all_clients)
			}
		case lost_orders := <-lost_orders_c:
			Search_for_lost_orders(lost_orders, order_from_cost, all_clients)
		case set_light := <-set_light_c:
			if set_light.Flag {
				//Send ligth request to controller
			} else {
				//Send ligth request to controller
			}
		case delete_order := <-del_order_c:
			network_list[delete_order.Button][delete_order.Floor] = false
			order_complete_c <- delete_order
		case send_order := <-order_from_cost:
			if send_order.Ip_from_cost.String() == localIP.String() {
				fmt.Println("I will handle this order \n \n")
				order_from_network <- send_order
			}
		}
	}
}

func Read_msg(msg_from_network chan Client, set_light_c chan Lights, del_order_c chan Order, localIP net.IP, all_clients map[string]Client) {
	laddr, err_conv_ip_listen := net.ResolveUDPAddr("udp", ":20003")
	Check_error(err_conv_ip_listen)
	listener, err_listen := net.ListenUDP("udp", laddr)
	Check_error(err_listen)
	var decoded_client Client
	var decoded_lights Lights
	var decoded_order Order
	for {
		b := make([]byte, 1024)
		n, _, _ := listener.ReadFromUDP(b)
		code := string(b[:3])
		switch code {
		case "cli":
			err_decoding := json.Unmarshal(b[3:n], &decoded_client)
			if err_decoding != nil {
				fmt.Println("error decoding client msg")
			}
			msg_from_network <- decoded_client

		case "lig":
			err_decoding := json.Unmarshal(b[3:n], &decoded_lights)
			if err_decoding == nil {
				set_light_c <- decoded_lights
			}
		case "del":
			err_decoding := json.Unmarshal(b[3:n], &decoded_order)
			if err_decoding == nil {
				del_order_c <- decoded_order
			}
		}
	}
}

func Send_msg(order_to_network chan Client, send_lights_c chan Lights, send_del_req_c chan Order) {
	baddr, err_conv_ip := net.ResolveUDPAddr("udp", "129.241.187.255:20003")
	Check_error(err_conv_ip)
	msg_sender, err_dialudp := net.DialUDP("udp", nil, baddr)
	Check_error(err_dialudp)
	for {
		select {
		case new_order := <-order_to_network:
			msg_encoded, err_encoding := json.Marshal(new_order)
			if err_encoding != nil {
				fmt.Println("error encoding json: ", err_encoding)
			}
			msg_encoded = append([]byte("cli"), msg_encoded...)
			msg_sender.Write(msg_encoded)
		case send_lights := <-send_lights_c:
			lights_encoded, err_encoding := json.Marshal(send_lights)
			if err_encoding != nil {
				fmt.Println("error encoding json: ", err_encoding)
			}
			lights_encoded = append([]byte("lig"), lights_encoded...)
			msg_sender.Write(lights_encoded)
		case send_del_req := <-send_del_req_c:
			delete_encoded, err_encoding := json.Marshal(send_del_req)
			if err_encoding != nil {
				fmt.Println("error encoding json: ", err_encoding)
			}
			delete_encoded = append([]byte("del"), delete_encoded...)
			msg_sender.Write(delete_encoded)
		}
	}
}

func Send_status(status_update_c chan Client) {
	baddr, err_conv_ip := net.ResolveUDPAddr("udp", "129.241.187.255:20020")
	Check_error(err_conv_ip)
	status_sender, err_dialudp := net.DialUDP("udp", nil, baddr)
	Check_error(err_dialudp)
	for {
		select {
		case status_update := <-status_update_c:
			status_encoded, err_encoding := json.Marshal(status_update)
			if err_encoding != nil {
				fmt.Println("error encoding json: ", err_encoding)
			}
			status_encoded = append([]byte("status"), status_encoded...)
			status_sender.Write([]byte(status_encoded))
		}

	}
}

func Read_status(lost_orders_c chan Client, all_ips map[string]time.Time, all_clients map[string]Client, localIP net.IP) {
	laddr, err_conv_ip_listen := net.ResolveUDPAddr("udp", ":20020")
	Check_error(err_conv_ip_listen)
	status_receiver, err_listen := net.ListenUDP("udp", laddr)
	Check_error(err_listen)
	var status_decoded Client
	for {
		time.Sleep(25 * time.Millisecond)
		b := make([]byte, 1024)
		n, raddr, _ := status_receiver.ReadFromUDP(b)
		code := string(b[:6])
		switch code {
		case "status":
			err_decoding := json.Unmarshal(b[6:n], &status_decoded)
			if err_decoding != nil {
				fmt.Println("error decoding client msg")
			}
			status_decoded.Ip = raddr.IP
			all_ips[raddr.IP.String()] = time.Now()
			all_clients[raddr.IP.String()] = status_decoded
			Write_to_file(status_decoded)
		}
		terminated, trm_client := CheckForElapsedClients(all_ips, all_clients)
		if terminated {
			okay, lost_client := Restore_floorpanel_orders(trm_client.Ip)
			if okay {
				lost_orders_c <- lost_client
			}
		}
	}
}

func CheckForElapsedClients(all_ips map[string]time.Time, all_clients map[string]Client) (bool, Client) {
	var client Client
	for key, value := range all_ips {
		if time.Now().Sub(value) > 2*time.Second {
			fmt.Println("Deleting IP: ", key, " ", value)
			client = all_clients[key]
			delete(all_ips, key)
			delete(all_clients, key)
			return true, client
		}
	}
	return false, client
}

func LocalIP() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address
				continue
			}
			return v4, nil
		}
	}
	return nil, nil //errors.New("cannot find local IP address")
}
