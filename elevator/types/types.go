package types

import (
	"net"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type Elev_button_type_t int

const (
	BUTTON_CALL_UP Elev_button_type_t = iota
	BUTTON_CALL_DOWN
	BUTTON_COMMAND
)

type Client struct {
	Floor         int
	Direction     int
	Current_floor int
	Cost          int
	Ip            net.IP
	Ip_from_cost  net.IP
	State         State_t
	NetState      NetState_t
	Order_list    [3][4]bool
	Button        Elev_button_type_t
}

type Order struct {
	Floor  int
	Dir    int
	Button Elev_button_type_t
}

type Lights struct {
	Floor  int
	Button Elev_button_type_t
	Flag   bool
}

type NetState_t int

const (
	OFF NetState_t = iota
	ON
)

type State_t int

const (
	RUN State_t = iota
	DOOR
	WAIT
	STOPS
	STOP_OBS
	UNDEFS
)

type Event_t int

const (
	NEW_ORDER Event_t = iota
	NO_ORDERS
	FLOOR_REACHED
	OBSTRUCTION
	STOP
	UNDEF
)
