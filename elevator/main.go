package main

import (
	network "./network"
	controller "./controller"
)

func main() {
	all_ips_m := make(map[string]time.Time)
	all_clients_m := make(map[string]driver.Client)

	msg_from_network := make(chan driver.Client, 1)
	order_to_network := make(chan driver.Client, 1)
	order_from_network := make(chan driver.Client, 3)
	order_from_cost := make(chan driver.Client, 3)
	status_update_c := make(chan driver.Client, 1)
	check_backup_c := make(chan driver.Client, 1)
	lost_orders_c := make(chan driver.Client, 1)
	send_lights_c := make(chan driver.Lights, 1)
	set_light_c := make(chan driver.Lights, 1)
	send_del_req_c := make(chan driver.Order, 1)
	del_order_c := make(chan driver.Order, 1)
	order_complete_c := make(chan driver.Order, 1)
	disconnected := make(chan int, 1)
	netstate_c := make(chan driver.NetState_t, 1)

	localIP, _ := LocalIP()
	fmt.Println(localIP)

	init_elevator, init_hardware, current_floor := controller.Elevator_init()
	if init_elevator && init_hardware {
		go controller.ControlHandler(order_from_network, order_to_network, check_backup_c, status_update_c, send_lights_c, send_del_req_c, order_complete_c, disconnected, netstate_c, current_floor, localIP)
	}

	restore_ok := Restore_command_orders(check_backup_c, localIP)
	if !restore_ok {
		fmt.Println("No orders to restore")
	}

	// Initilize main threads for network communication
	go network.Inter_process_communication(msg_from_network, order_from_network, order_from_cost, lost_orders_c, set_light_c, del_order_c, localIP, all_clients_m, order_complete_c)
	go network.Read_msg(msg_from_network, set_light_c, del_order_c, localIP, all_clients_m)
	go network.Send_msg(order_to_network, send_lights_c, send_del_req_c)
	go network.Read_status(lost_orders_c, all_ips_m, all_clients_m, localIP)
	go network.Send_status(status_update_c)
	go network.Get_kill_sig()
	go network.Check_connectivity(disconnected, netstate_c, all_clients_m, localIP)

	neverQuit := make(chan string)
	<-neverQuit
}
