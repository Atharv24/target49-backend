package udp_server

const (
	// S;{POS}:{TIMESTAMP} from client, S;{ID}:{POS}:{TIMESTAMP};{ID}:{POS}:{TIMESTAMP} from server
	PLAYER_STATE_MESSAGE = "S"

	// H;{HIT_PLAYER_ID}
	PLAYER_SHOT_MESSAGE = "H"

	// L;{NAME} from client
	PLAYER_LOGIN_MESSAGE = "L"

	// I;{NEW_PLAYER_ID}:{NEW_POS}:{TIMESTAMP};{ID1}:{POS1}:{TIMESTAMP};{ID2}:{POS2}:{TIMESTAMP} from server to the new client
	INITIAL_MESSAGE = "I"

	// N;{NEW_PLAYER_ID}:{NEW_POS}:{TIMESTAMP} from server to all existing clients
	NEW_PLAYER_MESSAGE = "N"

	// R;{POS}:{HEALTH}
	PLAYER_RESET_MESSAGE = "R"

	// P;{ID1}:{SCORE}:{DEATHS};{ID2}:{SCORE}:{DEATHS}
	POINTS_MESSAGE = "P"
)

const MAX_HEALTH = 5

const RESPAWN_IDLE_DELAY_MS = 2 * 1000 // 2 seconds
