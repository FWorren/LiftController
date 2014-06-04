package controller

import (
	"../types" 
	driver "../driver"
	"fmt"
	"time"
)

func Elevator_init() (init_elevator bool, init_hardware bool, prev Order) {
	init_hardware = true
	if Elev_init() == 0 {
		init_hardware = false
		fmt.Println("Unable to initialize elevator hardware\n")
	}
	fmt.Println("Press STOP button to stop elevator and exit program.\n")
	Elevator_clear_all_lights()
	Elev_set_speed(-300)
	var current Order
	for {
		time.Sleep(10 * time.Millisecond)
		floor := driver.Elev_get_floor_sensor_signal()
		if floor != -1 {
			driver.Elev_set_floor_indicator(floor)
			Elevator_break(-1)
			init_elevator = true
			current.Floor = floor
			current.Dir = -1
		}
	}
	return init_elevator, init_hardware, current_floor
}

func ElevatorHandler(head_order_c chan Order, prev_order_c chan Order, del_order chan Order, state_c chan State_t, update_local_list_c chan[N_BUTTONS][N_FLOORS]bool) {
	get_prev_floor_c := make(chan Order, 1)
	delete_order_c := make(chan Order, 1)
	state_update_c := make(chan State_t, 1)
	var local_list [N_BUTTONS][N_FLOORS]bool
	var head_order Order
	var prev_order Order

	go func () {
		for {
			select {
			case local_list = <- update_local_list_c:
			case head_order = <-head_order_c:
			case prev_order = <-get_prev_floor_c:
				prev_order_c <- prev_order
			case del_req := <-delete_order:
				del_order <- del_req
			case new_state := <-state:
				state_c <- new_state
			}
		}
	}();

	go func () {
		var state State_t
		var event Event_t
		for {
			time.Sleep(10 * time.Millisecond)
			switch event {
			case NEW_ORDER:
				event = Elevator_run(state_update_c, get_prev_floor_c, &state, head_order)
			case NO_ORDERS:
				event = Elevator_wait(state_update_c, &state)
			case FLOOR_REACHED:
				event = Elevator_door(state_update_c, &state)
			case OBSTRUCTION:
				event = Elevator_stop_obstruction(state_update_c, &state)
			case STOP:
				event = Elevator_stop(state_update_c, &state)
			case UNDEF:
				event = Elevator_undef(state_update_c, &state)
			}
		}
	}();
}

func Elvator_wait(state_update_c chan State_t, state *State_t) {
	if *state != WAIT {
		*state = WAIT
		state_update_c <- RUN
	}
	return NO_ORDERS
}

func Elevator_run(state_update_c chan State_t, get_prev_floor_c chan Order, state *State_t, head_order Order) {
	if *state != RUN {
		*state = RUN
		driver.Elev_set_speed(300 * head_order.Dir)
		state_update_c <- RUN
	}
	current_floor := driver.Elev_get_floor_sensor_signal()
	if current_floor != -1 {
		var current Order
		current.Floor = current_floor
		current.Dir = head_order.Dir
		get_prev_floor_c <- current
		driver.Elev_set_floor_indicator(current_floor)
		if current_floor == head_order.Floor {
			Elevator_break(head_order.dir)
			return FLOOR_REACHED
		}
	}
	if current_floor == head_order.Floor {
		Elevator_break(head_order.Dir)
		return FLOOR_REACHED
	}
	if driver.Elev_get_stop_signal() {
		Elevator_break(head_order.Dir)
		return STOP
	}
	if driver.Elev_get_obstruction_signal() {
		Elevator_break(head_order.Dir)
		return OBSTRUCTION
	}
	return NEW_ORDER
}

func Elevator_door(state_update_c chan State_t, state *State_t) {
	if driver.Elev_get_floor_sensor_signal() != -1 {
		if *state != DOOR {
			*state = DOOR
			state_update_c <- DOOR
			driver.Elev_set_door_open_lamp(1)
		}
		time.Sleep(3000 * time.Millisecond)
		if driver.Elev_get_obstruction_signal() {
			return DOOR
		}
		if driver.Elev_get_stop_signal() {
			return STOP
		}
		driver.Elev_set_door_open_lamp(0)
		return NEW_ORDER
	}else {
		return UNDEF
	}
}

func Elevator_stop(state_update_c chan State_t, state *State_t) {
	if *state != STOPS {
		*state = STOPS
		state_update_c <- STOPS
		Elevator_clear_all_lights()
		driver.Elev_set_stop_lamp(1)
		fmt.Println("The elevator has stopped!\n1. If you wish to order a new floor, do so, or.\n2. Press Ctrl + c to exit program.\n")	
	}
	return STOP
}

func Elevator_stop_obstruction(state_update_c chan State_t, state *State_t) {
	if *state != STOP_OBS {
		*state = STOP_OBS
		state_update_c <- STOP_OBS
	}
	if !driver.Elev_get_obstruction_signal() {
		return NEW_ORDER
	}
	return OBSTRUCTION
}

func Elevator_undef(state_update_c chan State_t, state *State_t) {
	if *state != UNDEFS {
		*state = UNDEFS
		state_update_c <- UNDEFS
	}
	return UNDEF
}

func Elevator_clear_all_lights() {
	driver.Elev_set_door_open_lamp(0)
	driver.Elev_set_stop_lamp(0)
	for i := 0; i < N_FLOORS; i++ {
		driver.Elev_set_button_lamp(BUTTON_COMMAND, i, 0)
		if i > 0 {
			driver.Elev_set_button_lamp(BUTTON_CALL_DOWN, i, 0)
		}
		if i < N_FLOORS-1 {
			driver.Elev_set_button_lamp(BUTTON_CALL_UP, i, 0)
		}
	}
}

func Elevator_break(direction int) {
	driver.Elev_set_speed(100 * (-direction))
	time.Sleep(20 * time.Millisecond)
	driver.Elev_set_speed(0)
}
