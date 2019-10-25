package metrics

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	StatusOk  Status = "ok"
	StatusErr Status = "err"
)
