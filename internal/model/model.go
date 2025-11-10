package model

// это так выглядит подобие енамов в го :)
// https://dizzy.zone/2024/01/26/Enums-in-Go/
type Status string

const (
	StatusOk    Status = "ok"
	StatusNotOk Status = "not_ok"
)
