package driver

var lamp_channel_matrix = [N_FLOORS][N_BUTTONS]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var button_channel_matrix = [N_FLOORS][N_BUTTONS]int{
	{FLOOR_UP1, FLOOR_DOWN1, FLOOR_COMMAND1},
	{FLOOR_UP2, FLOOR_DOWN2, FLOOR_COMMAND2},
	{FLOOR_UP3, FLOOR_DOWN3, FLOOR_COMMAND3},
	{FLOOR_UP4, FLOOR_DOWN4, FLOOR_COMMAND4},
}