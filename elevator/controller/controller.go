package controller

import (
	"../types"
	driver "../driver"
	"net"
	"time"
)

func ControlHandler(order_from_network chan Client, order_to_network chan Client, check_backup_c chan Client, status_update_c chan Client, send_lights_c chan Lights, send_del_req_c chan Order, order_complete_c chan Order, disconnected chan int, netstate_c chan NetState_t, current_floor Order, localIP net.IP) {
	order_internal := make(chan Order, 1)
	head_order_c := make(chan Order, 1)
	prev_order_c := make(chan Order, 1)
	del_Order := make(chan Order, 1)
	reset_list_c := make(chan Order, 1)
	state_c := make(chan State_t, 1)
	reset_all_c := make(chan int, 1)
	convenient_list_c := make(chan [N_BUTTONS][N_FLOORS]bool, 1)

	var state State_t
	var netState NetState_t
	var local_list [N_BUTTONS][N_FLOORS]bool
	var client Client
	var Head_order Order
	var light Lights
	Prev_order := current_floor
	client.Current_floor = current_floor.Floor
	client.Direction = current_floor.Dir

	go ElevatorHandler(head_order_c, prev_order_c, del_Order, state_c, convenient_list_c)
	go Search_for_orders(order_internal, reset_list_c, reset_all_c)

	timeOut := make(<-chan time.Time)
	for {
		timeOut = time.After(100 * time.Millisecond)
		select {
		case to_network := <-order_internal:
			client.Floor = to_network.Floor
			client.Button = to_network.Button
			client.Ip = localIP
			client.Current_floor = Prev_order.Floor
			if to_network.Button == BUTTON_COMMAND && !local_list[BUTTON_COMMAND][to_network.Floor] {
				local_list[BUTTON_COMMAND][to_network.Floor] = true
				client.Order_list[BUTTON_COMMAND][to_network.Floor] = true
			} else {
				if netState == ON {
					order_to_network <- client
				} else {
					reset_list_c <- to_network
				}
			}

		case from_network := <-order_from_network:
			local_list[from_network.Button][from_network.Floor] = true
			client.Order_list[from_network.Button][from_network.Floor] = true
			light.Floor = from_network.Floor
			light.Button = from_network.Button
			light.Flag = true
			send_lights_c <- light
		case backup_client := <-check_backup_c:
			client.Order_list = Get_backup_orders(backup_client)
			local_list = client.Order_list
		case state = <-state_c:
			client.State = state
		case Update_prev := <-prev_order_c:
			Prev_order = Update_prev
			client.Direction = Prev_order.Dir
			client.Current_floor = Prev_order.Floor
			convenient_list_c <- local_list
		case del_msg := <-del_Order:
			local_list[del_msg.Button][del_msg.Floor] = false
			client.Order_list[del_msg.Button][del_msg.Floor] = false
			reset_list_c <- del_msg
			light.Floor = del_msg.Floor
			light.Button = del_msg.Button
			light.Flag = false
			send_lights_c <- light
			send_del_req_c <- del_msg
		case completed_order := <-order_complete_c:
			reset_list_c <- completed_order
		case netState = <-netstate_c:
			client.NetState = netState
		case <-disconnected:
			Delete_all_orders(&local_list);
			reset_all_c <- 1
		case <-timeOut:
			has_order := Check_number_of_local_orders(local_list)
			if (state == WAIT || state == UNDEF) && has_order {
				Head_order = Set_head_order(local_list, Head_order, Prev_order)
				client.Direction = Head_order.Dir
				head_order_c <- Head_order
			}
			status_update_c <- client
		}
	}
}
