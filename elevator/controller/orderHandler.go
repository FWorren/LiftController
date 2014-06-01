package controller

import (
	types "../types"
	driver "../driver"
	"net"
	"time"
)

func Search_for_orders(order_internal chan types.Order, reset_list_c chan types.Order, reset_all_c chan int) {
	var new_order types.Order
	var list [types.N_BUTTONS][types.N_FLOORS]bool

	go func() {
		for {
			select {
			case reset_floor := <-reset_list_c:
				list[reset_floor.Button][reset_floor.Floor] = false
			case <-reset_all_c:
				for i := 0; i < types.N_FLOORS; i++ {
					list[types.BUTTON_CALL_DOWN][i] = false
					list[types.BUTTON_CALL_UP][i] = false
				}
			}
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			for i := 0; i < types.N_FLOORS; i++ {
				if driver.Elev_get_button_signal(types.BUTTON_COMMAND, i) == 1 {
					if !list[types.BUTTON_COMMAND][i] {
						new_order.Button = types.BUTTON_COMMAND
						new_order.Floor = i
						list[types.BUTTON_COMMAND][i] = true
						Elev_set_button_lamp(types.BUTTON_COMMAND, i, 1)
						order_internal <- new_order
					}
				}
				if i > 0 {
					if driver.Elev_get_button_signal(types.BUTTON_CALL_DOWN, i) == 1 {
						if !list[types.BUTTON_CALL_DOWN][i] {
							new_order.Floor = i
							new_order.Button = types.BUTTON_CALL_DOWN
							list[types.BUTTON_CALL_DOWN][i] = true
							order_internal <- new_order
						}
					}
				}
				if i < types.N_FLOORS-1 {
					if driver.Elev_get_button_signal(types.BUTTON_CALL_UP, i) == 1 {
						if !list[types.BUTTON_CALL_UP][i] {
							new_order.Floor = i
							new_order.Button = types.BUTTON_CALL_UP
							list[types.BUTTON_CALL_UP][i] = true
							order_internal <- new_order
						}
					}
				}
			}
		}
	}()
}

func Get_backup_orders(client types.Client) [3][4]bool {
	var command_list [types.N_BUTTONS][types.N_FLOORS]bool
	for i := 0; i < types.N_FLOORS; i++ {
		if client.Order_list[types.BUTTON_COMMAND][i] {
			command_list[types.BUTTON_COMMAND][i] = true
			driver.Elev_set_button_lamp(types.BUTTON_COMMAND, i, 1)
		}
	}
	return command_list
}

func Check_number_of_local_orders(local_list [3][4]bool) bool {
	numb_orders := 0
	for i := 0; i < types.N_FLOORS; i++ {
		if local_list[types.BUTTON_CALL_UP][i] {
			numb_orders++
		}
		if local_list[types.BUTTON_CALL_DOWN][i] {
			numb_orders++
		}
		if local_list[types.BUTTON_COMMAND][i] {
			numb_orders++
		}
	}
	if numb_orders > 0 {
		return true
	} else {
		return false
	}
}

func Set_head_order(local_list [3][4]bool, Head_order types.Order, Prev_order types.Order) types.Order {
	switch Prev_order.Dir {
	case 1:
		new_head := OrderHandler_state_up(local_list, Head_order, Prev_order)
		if new_head.Floor != -1 {
			Head_order = new_head
			return Head_order
		}
		Prev_order.Dir = new_head.Dir
	case -1:
		new_head := OrderHandler_state_down(local_list, Head_order, Prev_order)
		if new_head.Floor != -1 {
			Head_order = new_head
			return Head_order
		}
		Prev_order.Dir = new_head.Dir
	}
}

func State_up(local_list [3][4]bool, Head_order types.Order, Prev_order types.Order) types.Order {
	if Prev_order.Floor == N_FLOORS-1 {
		Head_order.Dir = -1
		Head_order.Floor = -1
		return Head_order
	}
	for i := Prev_order.Floor; i < types.N_FLOORS; i++ {
		if local_list[types.BUTTON_CALL_UP][i] {
			Head_order.Floor = i
			Head_order.Dir = 1
			Head_order.Button = types.BUTTON_CALL_UP
			return Head_order
		}
		if local_list[types.BUTTON_CALL_DOWN][i] {
			Head_order.Floor = i
			Head_order.Dir = 1
			Head_order.Button = types.BUTTON_CALL_DOWN
			return Head_order
		}
		if local_list[types.BUTTON_COMMAND][i] {
			Head_order.Floor = i
			Head_order.Dir = 1
			Head_order.Button = types.BUTTON_COMMAND
			return Head_order
		}
	}
	Head_order.Floor = -1
	Head_order.Dir = -1
	return Head_order
}

func State_down(local_list [3][4]bool, Head_order types.Order, Prev_order types.Order) types.Order {
	if Prev_order.Floor == 0 {
		Head_order.Dir = 1
		Head_order.Floor = -1
		return Head_order
	}
	for i := Prev_order.Floor; i >= 0; i-- {
		if local_list[types.BUTTON_CALL_UP][i] {
			Head_order.Floor = i
			Head_order.Dir = -1
			Head_order.Button = types.BUTTON_CALL_UP
			return Head_order
		}
		if local_list[types.BUTTON_CALL_DOWN][i] {
			Head_order.Floor = i
			Head_order.Dir = -1
			Head_order.Button = types.BUTTON_CALL_DOWN
			return Head_order
		}
		if local_list[types.BUTTON_COMMAND][i] {
			Head_order.Floor = i
			Head_order.Dir = -1
			Head_order.Button = types.BUTTON_COMMAND
			return Head_order
		}
	}
	Head_order.Floor = -1
	Head_order.Dir = 1
	return Head_order
}

func Delete_all_orders(local_list *[3][4]bool) {
	for i := 0; i < types.N_FLOORS; i++ {
		local_list[types.BUTTON_CALL_DOWN][i] = false
		local_list[types.BUTTON_CALL_UP][i] = false
		client.Order_list[types.BUTTON_CALL_DOWN][i] = false
		client.Order_list[types.BUTTON_CALL_UP][i] = false
		driver.Elev_set_button_lamp(types.BUTTON_CALL_DOWN, i, 0)
		driver.Elev_set_button_lamp(types.BUTTON_CALL_UP, i, 0)
	}
}