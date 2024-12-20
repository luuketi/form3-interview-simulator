package payment

import (
	"github.com/form3tech-oss/interview-simulator/internal/response"
	"strconv"
	"strings"
	"time"
)

type Payment struct {
	Amount      uint64
	ErrorReason string
}

func FromString(request string) Payment {
	parts := strings.Split(request, "|")
	if len(parts) != 2 || parts[0] != "PAYMENT" {
		return Payment{ErrorReason: "Invalid request"}
	}

	amount, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return Payment{ErrorReason: "Invalid amount"}
	}
	return Payment{Amount: amount}
}

func (p Payment) Process() response.Response {
	if p.ErrorReason != "" {
		return response.NewRejected(p.ErrorReason)
	}

	if p.Amount > 100 {
		processingTime := p.Amount
		if p.Amount > 10000 {
			processingTime = 10000
		}
		time.Sleep(time.Duration(processingTime) * time.Millisecond)
	}
	return response.NewAccepted("Transaction processed")
}
