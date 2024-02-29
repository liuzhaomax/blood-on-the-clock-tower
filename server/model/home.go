package model

type HomeReqBody struct {
	Action  string `json:"action"`
	Payload any    `json:"payload"`
}

type ListRoomsReqBody struct {
	Action  string `json:"action"`
	Payload string `json:"payload"`
}

type CreateRoomReqBody struct {
	Action  string `json:"action"`
	Payload Room   `json:"payload"`
}

type JoinRoomReqBody struct {
	Action  string          `json:"action"`
	Payload JoinRoomPayload `json:"payload"`
}

type JoinRoomPayload struct {
	Room   RoomInJoinRoomPayload `json:"room"`
	Player Player                `json:"player"`
}

type RoomInJoinRoomPayload struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}
