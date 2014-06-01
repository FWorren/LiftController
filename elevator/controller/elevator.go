package controller

import (
	driver "../driver"
	"fmt"
	"time"
)

func Elevator_init() (init_elevator bool, init_hardware bool, prev driver.Order) {
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
		floor := Elev_get_floor_sensor_signal()
		if floor != -1 {
			Elev_set_floor_indicator(floor)
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
	state_c := make(chan State_t, 1)
	update_local_list := make(chan [N_BUTTONS][N_FLOORS]bool)
	var head_order Order
	var prev_order Order

	go func () {
		for {
			select {
			case update_local_list =<- :
			case head_order =<-head_order_c:
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
		for {
			time.Sleep(10 * time.Millisecond)
			switch event {
			case NEW_ORDER:
				event = Elevator_run()
			case NO_ORDER:
				event = Elevator_wait()
			case FLOOR_REACHED:
				event = Elevator_door()
			case OBSTRUCTION:
				event = Elevator_stop_obstruction()
			case STOP:
				event = Elevator_stop()
			case UNDEF:
				event = Elevator_undef()
			}
		}
	}();
}

func Elvator_wait() {
	
}

func Elevator_run() {
	Elev_set_speed(300 * head_order.Dir)
	current_floor := Elev_get_floor_sensor_signal()
	if current_floor != -1 {
		var current Order
		current.Floor = current_floor
		current.Dir = head_order.Dir
		get_prev_floor_c <- current
		Elev_set_floor_indicator(current_floor)
	}
	if current_floor == head_order.Floor {
		Elevator_break(head_order.Dir)
		floor_reached <- head_order
		return FLOOR_REACHED
	}
	if Elev_get_stop_signal() {
		Elevator_break(head_order.Dir)
		return STOP
	}
	if Elev_get_obstruction_signal() {
		Elevator_break(head_order.Dir)
		return OBSTRUCTION
	}
}

func Elevator_door(head_order Order, delete_order chan Order, state chan State_t) {
	if Elev_get_floor_sensor_signal() != -1 {
		Elev_set_door_open_lamp(1)
	}else {
		return UNDEF
	}
}

func Elevator_stop(state chan State_t) {
	state <- STOPS
	fmt.Println("The elevator has stopped!\n1. If you wish to order a new floor, do so, or.\n2. Press Ctrl + c to exit program.\n")
	Elevator_clear_all_lights()
	Elev_set_stop_lamp(1)
}

func Elevator_stop_obstruction(head_order_c chan Order, head_order Order, state chan State_t) {
	
}

func Elevator_clear_all_lights() {
	Elev_set_door_open_lamp(0)
	Elev_set_stop_lamp(0)
	for i := 0; i < N_FLOORS; i++ {
		Elev_set_button_lamp(BUTTON_COMMAND, i, 0)
		if i > 0 {
			Elev_set_button_lamp(BUTTON_CALL_DOWN, i, 0)
		}
		if i < N_FLOORS-1 {
			Elev_set_button_lamp(BUTTON_CALL_UP, i, 0)
		}
	}
}

func Elevator_break(direction int) {
	Elev_set_speed(100 * (-direction))
	time.Sleep(20 * time.Millisecond)
	Elev_set_speed(0)
}
