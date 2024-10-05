package model

type Uploadstatus struct {
	BucketName string
	ObjectName string
	Status     bool
}

type ResponseModel struct {
	Status      bool
	Status_code int
	Data        interface{}
	Error       interface{}
}

type Files struct {
	FilePath string
}
