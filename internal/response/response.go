package response

import "fmt"

type Response struct {
	Status string
	Reason string
}

func NewAccepted(reason string) Response {
	return Response{"ACCEPTED", reason}
}

func NewRejected(reason string) Response {
	return Response{"REJECTED", reason}
}

func (r *Response) ToString() string {
	return fmt.Sprintf("RESPONSE|%s|%s", r.Status, r.Reason)
}
